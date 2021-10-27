package worker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/alancesar/photo-gallery/thumbs/image"
	"io"
	"path/filepath"
	"strings"
	"sync"
)

type Storage interface {
	Put(ctx context.Context, img image.Image) error
	Get(ctx context.Context, filename string) (io.Reader, error)
}

type Database interface {
	Create(filename string, img image.Image) error
	FindByFilenameAndType(filename string, imageType image.Type) ([]image.Image, bool, error)
}

type Processor interface {
	Fit(reader io.Reader, dimension image.Dimension) (io.Reader, image.Dimension, error)
}

type Producer interface {
	Produce(filename string, images []image.Image) error
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

func (t Thumbs) CreateThumbnails(ctx context.Context, filename string, metadata image.Metadata) error {
	raw, err := t.getRawData(ctx, filename)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(len(t.dimensions))

	worker := func(reader io.Reader, targetDimension image.Dimension) {
		defer wg.Done()

		resized, realDimension, err := t.processor.Fit(reader, targetDimension)
		if err != nil {
			fmt.Println(err)
			return
		}

		img := image.Image{
			Reader:   resized,
			Filename: createFilename(filename, realDimension),
			Type:     image.Thumb,
			Metadata: image.Metadata{
				ContentType: metadata.ContentType,
				Dimension:   realDimension,
			},
		}

		if err := t.thumbStorage.Put(ctx, img); err != nil {
			fmt.Println(err)
			return
		}

		if err := t.database.Create(filename, img); err != nil {
			fmt.Println(err)
			return
		}
	}

	for _, dimension := range t.dimensions {
		go worker(io.NopCloser(bytes.NewBuffer(raw)), dimension)
	}

	wg.Wait()

	thumbs, err := t.getThumbsFromDatabase(filename)
	if err != nil {
		return err
	}

	return t.producer.Produce(filename, thumbs)
}

func (t Thumbs) getRawData(ctx context.Context, filename string) ([]byte, error) {
	reader, err := t.photoStorage.Get(ctx, filename)
	if err != nil {
		return nil, err
	}

	raw, err := io.ReadAll(reader)
	return raw, err
}

func (t Thumbs) getThumbsFromDatabase(filename string) ([]image.Image, error) {
	thumbs, exists, err := t.database.FindByFilenameAndType(filename, image.Thumb)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, errors.New(fmt.Sprintf("Could not find thumbs for %s", filename))
	}

	return thumbs, err
}

func createFilename(filename string, dimension image.Dimension) string {
	ext := filepath.Ext(filename)
	filename = strings.TrimSuffix(filename, ext)
	return fmt.Sprintf("%s_%dx%d%s", filename, dimension.Width, dimension.Height, ext)
}
