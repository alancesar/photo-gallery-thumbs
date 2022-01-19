package worker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/alancesar/photo-gallery/thumbs/image"
	"io"
	"path/filepath"
	"strings"
	"sync"
)

const (
	jpegExtension   = ".jpeg"
	jpegContentType = "image/jpeg"
	thumbsProperty  = "thumbs"
)

type Storage interface {
	Put(ctx context.Context, img image.Image) error
	Get(ctx context.Context, filename string) (io.ReadSeeker, error)
}

type Database interface {
	Create(filename string, img image.Image) error
	FindByFilenameAndType(filename string, imageType image.Type) ([]image.Image, bool, error)
}

type Processor interface {
	Fit(reader io.Reader, dimension image.Dimension) (io.Reader, image.Dimension, error)
}

type Producer interface {
	Produce(message image.Message) error
}

type Thumbs struct {
	photoStorage Storage
	thumbStorage Storage
	database     Database
	processor    Processor
	producer     Producer
	dimensions   []image.Dimension
}

type Bundle struct {
	PhotoStorage Storage
	ThumbStorage Storage
	Database     Database
	Processor    Processor
	Producer     Producer
	Dimensions   []image.Dimension
}

func NewThumbsWorker(bundle Bundle) *Thumbs {
	return &Thumbs{
		photoStorage: bundle.PhotoStorage,
		thumbStorage: bundle.ThumbStorage,
		database:     bundle.Database,
		processor:    bundle.Processor,
		producer:     bundle.Producer,
		dimensions:   bundle.Dimensions,
	}
}

func (t Thumbs) CreateThumbnails(ctx context.Context, imgFilename string) error {
	sample, err := t.getSampleImg(ctx, imgFilename)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(len(t.dimensions))
	mutex := sync.Mutex{}

	worker := func(item io.ReadSeeker, dimension image.Dimension) {
		defer wg.Done()

		mutex.Lock()
		thumb, thumbDimension, err := t.resize(item, dimension)
		if err != nil {
			fmt.Println(err)
			return
		}
		mutex.Unlock()

		thumbFilename := createThumbFilename(imgFilename, thumbDimension)
		img := createImgFromThumb(thumb, thumbFilename, thumbDimension)
		if err := t.saveImg(ctx, imgFilename, img); err != nil {
			fmt.Println(err)
			return
		}
	}

	for _, dimension := range t.dimensions {
		go worker(sample, dimension)
	}

	wg.Wait()

	thumbs, err := t.getThumbsFromDatabase(imgFilename)
	if err != nil {
		return err
	}

	return t.producer.Produce(image.Message{
		Filename: imgFilename,
		Property: thumbsProperty,
		Payload:  thumbs,
	})
}

func (t Thumbs) getSampleImg(ctx context.Context, filename string) (io.ReadSeeker, error) {
	item, err := t.photoStorage.Get(ctx, filename)
	if err != nil {
		return nil, err
	}

	largest := t.getLargestDimension()
	return t.createSampleImg(item, largest)
}

func (t Thumbs) getLargestDimension() image.Dimension {
	largest := image.Dimension{}
	for _, dimension := range t.dimensions {
		if dimension.Width > largest.Width {
			largest = dimension
		}
	}

	return largest
}

func (t Thumbs) createSampleImg(input io.Reader, dimension image.Dimension) (io.ReadSeeker, error) {
	sample, _, err := t.processor.Fit(input, dimension)
	if err != nil {
		return nil, err
	}

	content, err := io.ReadAll(sample)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(content), err
}

func (t Thumbs) resize(item io.ReadSeeker, dimension image.Dimension) (io.Reader, image.Dimension, error) {
	if _, err := item.Seek(0, 0); err != nil {
		return nil, image.Dimension{}, err
	}

	return t.processor.Fit(item, dimension)
}

func (t Thumbs) getThumbsFromDatabase(filename string) ([]image.Image, error) {
	thumbs, exists, err := t.database.FindByFilenameAndType(filename, image.Thumb)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, fmt.Errorf("could not find thumbs for %s", filename)
	}

	return thumbs, err
}

func (t Thumbs) saveImg(ctx context.Context, filename string, img image.Image) error {
	if err := t.thumbStorage.Put(ctx, img); err != nil {
		return err
	}

	return t.database.Create(filename, img)
}

func createImgFromThumb(reader io.Reader, filename string, dimension image.Dimension) image.Image {
	return image.Image{
		Reader:   reader,
		Filename: filename,
		Type:     image.Thumb,
		Metadata: image.Metadata{
			ContentType: jpegContentType,
			Dimension:   dimension,
		},
	}
}

func createThumbFilename(filename string, dimension image.Dimension) string {
	ext := filepath.Ext(filename)
	filename = strings.TrimSuffix(filename, ext)
	return fmt.Sprintf("%s_%dx%d%s", filename, dimension.Width, dimension.Height, jpegExtension)
}
