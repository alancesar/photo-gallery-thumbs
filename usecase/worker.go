package usecase

import (
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

type (
	Storage interface {
		Put(ctx context.Context, img image.Image) error
		Get(ctx context.Context, filename string) (io.ReadSeeker, error)
	}

	Processor interface {
		Fit(reader io.Reader, dimension image.Dimension) (io.Reader, image.Dimension, error)
	}

	Publisher interface {
		Publish(ctx context.Context, thumbs thumbs.Thumbs, attributes map[string]string)
	}

	CreateThumbsRequest struct {
		Filename   string
		Metadata   image.Metadata
		Dimensions []image.Dimension
	}

	Thumbs struct {
		photoStorage  Storage
		thumbsStorage Storage
		processor     Processor
		publisher     Publisher
	}
)

func NewThumbs(photoStorage, thumbsStorage Storage, processor Processor, publisher Publisher) *Thumbs {
	return &Thumbs{
		photoStorage:  photoStorage,
		thumbsStorage: thumbsStorage,
		processor:     processor,
		publisher:     publisher,
	}
}

func (t Thumbs) CreateThumbnails(ctx context.Context, request CreateThumbsRequest) error {
	item, err := t.photoStorage.Get(ctx, request.Filename)
	if err != nil {
		return err
	}

	thumbnails := t.createThumbnails(item, request)

	t.putOnStorage(ctx, thumbnails)

	t.publisher.Publish(ctx, thumbs.Thumbs{
		Filename: request.Filename,
		Images:   thumbnails,
	}, map[string]string{"event-type": "WORKER"})

	return nil
}

func (t Thumbs) createThumbnails(item io.ReadSeeker, request CreateThumbsRequest) []image.Image {
	var thumbnails []image.Image
	for _, dimension := range request.Dimensions {
		_, err := item.Seek(0, 0)
		if err != nil {
			log.Println(err)
			continue
		}

		resized, realDimension, err := t.processor.Fit(item, dimension)
		if err != nil {
			log.Println(err)
			continue
		}

		thumbnails = append(thumbnails, t.buildThumb(resized, request, realDimension))
	}

	return thumbnails
}

func (t Thumbs) buildThumb(reader io.Reader, request CreateThumbsRequest, realDimension image.Dimension) image.Image {
	return image.Image{
		Reader:   reader,
		Filename: createFilename(request.Filename, realDimension),
		Type:     image.Thumb,
		Metadata: image.Metadata{
			ContentType: request.Metadata.ContentType,
			Dimension:   realDimension,
		},
	}
}

func (t Thumbs) putOnStorage(ctx context.Context, thumbnails []image.Image) {
	wg := sync.WaitGroup{}
	wg.Add(len(thumbnails))

	for _, thumb := range thumbnails {
		go func(img image.Image) {
			defer wg.Done()

			if err := t.thumbsStorage.Put(ctx, img); err != nil {
				fmt.Println(err)
				return
			}
		}(thumb)
	}

	wg.Wait()
}

func createFilename(filename string, dimension image.Dimension) string {
	ext := filepath.Ext(filename)
	filename = strings.TrimSuffix(filename, ext)
	return fmt.Sprintf("%s_%dx%d.jpeg", filename, dimension.Width, dimension.Height)
}
