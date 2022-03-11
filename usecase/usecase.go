package usecase

import (
	"bytes"
	"context"
	"fmt"
	"github.com/alancesar/photo-gallery/thumbs/domain/image"
	"github.com/alancesar/photo-gallery/thumbs/domain/thumbs"
	"io"
	"log"
	"path/filepath"
	"strings"
	"sync"
)

const (
	photosDirectory = "/photos"
	jpegExtension   = ".jpeg"
	jpegContentType = "image/jpeg"
	eventTypeKey    = "event-type"
	workerEventType = "WORKER"
)

type (
	Storage interface {
		Put(ctx context.Context, img image.Image) error
		Get(ctx context.Context, filename string) (io.Reader, error)
	}

	Processor interface {
		Fit(reader io.Reader, dimension image.Dimension) (io.Reader, image.Dimension, error)
	}

	Publisher interface {
		Publish(ctx context.Context, thumbs thumbs.Thumbs, attributes map[string]string)
	}

	Thumbs struct {
		storage   Storage
		processor Processor
		publisher Publisher
	}
)

func NewThumbs(storage Storage, processor Processor, publisher Publisher) *Thumbs {
	return &Thumbs{
		storage:   storage,
		processor: processor,
		publisher: publisher,
	}
}

func (t Thumbs) CreateThumbnails(ctx context.Context, filename string, dimensions []image.Dimension) error {
	sample, err := t.getSampleImg(ctx, filename, dimensions)
	if err != nil {
		return err
	}

	images := t.createImages(sample, filename, dimensions)
	t.putOnStorage(ctx, images)

	t.publisher.Publish(ctx, thumbs.Thumbs{
		Filename: filename,
		Images:   images,
	}, map[string]string{eventTypeKey: workerEventType})

	return nil
}

func (t Thumbs) getLargestDimension(dimensions []image.Dimension) image.Dimension {
	largest := image.Dimension{}
	for _, dimension := range dimensions {
		if dimension.Width > largest.Width {
			largest = dimension
		}
	}

	return largest
}

func (t Thumbs) getSampleImg(ctx context.Context, filename string, dimensions []image.Dimension) (io.ReadSeeker, error) {
	item, err := t.storage.Get(ctx, filename)
	if err != nil {
		return nil, err
	}

	largest := t.getLargestDimension(dimensions)
	return t.createSampleImg(item, largest)
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

func (t Thumbs) createImages(item io.ReadSeeker, filename string, dimensions []image.Dimension) []image.Image {
	var images []image.Image
	for _, dimension := range dimensions {
		thumb, err := t.createThumb(item, filename, dimension)
		if err != nil {
			log.Println(err)
			continue
		}

		images = append(images, thumb)
	}

	return images
}

func (t Thumbs) createThumb(seeker io.ReadSeeker, filename string, dimension image.Dimension) (image.Image, error) {
	_, err := seeker.Seek(0, 0)
	if err != nil {
		return image.Image{}, err
	}

	resized, realDimension, err := t.processor.Fit(seeker, dimension)
	if err != nil {
		return image.Image{}, err
	}

	return createImageFromThumb(resized, filename, realDimension), nil
}

func (t Thumbs) putOnStorage(ctx context.Context, thumbnails []image.Image) {
	wg := sync.WaitGroup{}
	wg.Add(len(thumbnails))

	for _, thumb := range thumbnails {
		go func(img image.Image) {
			defer wg.Done()

			if err := t.storage.Put(ctx, img); err != nil {
				fmt.Println(err)
				return
			}
		}(thumb)
	}

	wg.Wait()
}

func createImageFromThumb(reader io.Reader, filename string, realDimension image.Dimension) image.Image {
	return image.Image{
		Reader:   reader,
		Filename: createThumbFilename(filename, realDimension),
		Metadata: image.Metadata{
			ContentType: jpegContentType,
			Dimension:   realDimension,
		},
	}
}

func createThumbFilename(filename string, dimension image.Dimension) string {
	ext := filepath.Ext(filename)
	filename = strings.TrimSuffix(filename, ext)
	_, filename, _ = strings.Cut(filename, photosDirectory)
	return fmt.Sprintf("thumbs/%s_%dx%d%s", filename, dimension.Width, dimension.Height, jpegExtension)
}
