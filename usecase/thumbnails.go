package usecase

import (
	"bytes"
	"context"
	"fmt"
	"github.com/alancesar/photo-gallery/thumbs/domain/photo"
	"github.com/alancesar/photo-gallery/thumbs/presenter/message"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"path/filepath"
	"strings"
)

const (
	photosDirectory = "photos/"
	jpegExtension   = ".jpeg"
	jpegContentType = "image/jpeg"
)

type (
	Thumbs struct {
		storage   Storage
		processor Processor
		publisher Publisher
	}
)

func NewThumbnails(storage Storage, processor Processor, publisher Publisher) *Thumbs {
	return &Thumbs{
		storage:   storage,
		processor: processor,
		publisher: publisher,
	}
}

func (t Thumbs) Execute(ctx context.Context, id, filename string, dimensions []photo.Dimension) error {
	sample, err := t.getSampleImg(ctx, filename, dimensions)
	if err != nil {
		return err
	}

	images := t.createImages(sample, filename, dimensions)
	if err := t.putOnStorage(ctx, images); err != nil {
		return err
	}

	t.publisher.Publish(ctx, message.Photo{
		ID:     id,
		Thumbs: images,
	})

	return nil
}

func (t Thumbs) getLargestDimension(dimensions []photo.Dimension) photo.Dimension {
	largest := photo.Dimension{}
	for _, dimension := range dimensions {
		if dimension.Width > largest.Width {
			largest = dimension
		}
	}

	return largest
}

func (t Thumbs) getSampleImg(ctx context.Context, filename string, dimensions []photo.Dimension) (io.ReadSeeker, error) {
	item, err := t.storage.Get(ctx, filename)
	if err != nil {
		return nil, err
	}

	largest := t.getLargestDimension(dimensions)
	return t.createSampleImg(item, largest)
}

func (t Thumbs) createSampleImg(input io.Reader, dimension photo.Dimension) (io.ReadSeeker, error) {
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

func (t Thumbs) createImages(item io.ReadSeeker, filename string, dimensions []photo.Dimension) []photo.Image {
	var images []photo.Image
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

func (t Thumbs) createThumb(seeker io.ReadSeeker, filename string, dimension photo.Dimension) (photo.Image, error) {
	_, err := seeker.Seek(0, 0)
	if err != nil {
		return photo.Image{}, err
	}

	resized, realDimension, err := t.processor.Fit(seeker, dimension)
	if err != nil {
		return photo.Image{}, err
	}

	return createImageFromThumb(resized, filename, realDimension), nil
}

func (t Thumbs) putOnStorage(ctx context.Context, thumbnails []photo.Image) error {
	group, _ := errgroup.WithContext(ctx)
	for _, thumb := range thumbnails {
		worker := func() error {
			return t.storage.Put(ctx, thumb)
		}

		group.Go(worker)
	}

	return group.Wait()
}

func createImageFromThumb(reader io.Reader, filename string, realDimension photo.Dimension) photo.Image {
	return photo.Image{
		Reader:   reader,
		Filename: createThumbFilename(filename, realDimension),
		Metadata: photo.Metadata{
			ContentType: jpegContentType,
			Dimension:   realDimension,
		},
	}
}

func createThumbFilename(filename string, dimension photo.Dimension) string {
	ext := filepath.Ext(filename)
	filename = strings.TrimSuffix(filename, ext)
	filename = strings.TrimPrefix(filename, photosDirectory)
	return fmt.Sprintf("thumbs/%s_%dx%d%s", filename, dimension.Width, dimension.Height, jpegExtension)
}
