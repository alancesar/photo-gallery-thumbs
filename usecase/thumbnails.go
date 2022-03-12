package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/alancesar/photo-gallery/thumbs/domain/metadata"
	"github.com/alancesar/photo-gallery/thumbs/domain/thumb"
	"github.com/alancesar/photo-gallery/thumbs/pkg"
	"golang.org/x/sync/errgroup"
	"io"
	"path/filepath"
	"strings"
)

const (
	photosDirectory = "photos/"
	jpegExtension   = ".jpeg"
)

type (
	ThumbnailsWorker struct {
		storage   Storage
		processor Processor
	}
)

func NewThumbnails(storage Storage, processor Processor) *ThumbnailsWorker {
	return &ThumbnailsWorker{
		storage:   storage,
		processor: processor,
	}
}

func (t ThumbnailsWorker) Execute(ctx context.Context, filename string, dimensions []metadata.Dimension) ([]thumb.Thumbnail, error) {
	original, err := t.storage.Get(ctx, filename)
	if err != nil {
		return nil, err
	}

	largestDimension := metadata.GetLargestDimension(dimensions...)
	sample, err := t.processor.FitAsReadSeeker(original, largestDimension)
	if err != nil {
		return nil, err
	}

	thumbnails, err := t.generateThumbnails(sample, filename, dimensions)
	if err != nil {
		return nil, err
	}

	if err := t.putOnStorage(ctx, thumbnails); err != nil {
		return nil, err
	}

	return thumbnails, nil
}

func (t ThumbnailsWorker) generateThumbnails(seeker io.ReadSeeker, filename string, dimensions []metadata.Dimension) ([]thumb.Thumbnail, error) {
	var thumbnails []thumb.Thumbnail
	for _, dimension := range dimensions {
		resized, err := t.processor.FitFromReadSeeker(seeker, dimension)
		if err != nil {
			if errors.Is(err, pkg.ErrInvalidThumbSize) {
				continue
			}

			return nil, err
		}

		thumbnails = append(thumbnails, thumb.Thumbnail{
			Filename: createThumbFilename(filename, resized.Dimension),
			Image:    resized,
		})
	}

	return thumbnails, nil
}

func (t ThumbnailsWorker) putOnStorage(ctx context.Context, thumbnails []thumb.Thumbnail) error {
	group, _ := errgroup.WithContext(ctx)
	for _, thumbnail := range thumbnails {
		worker := func() error {
			return t.storage.Put(ctx, thumbnail)
		}

		group.Go(worker)
	}

	return group.Wait()
}

func createThumbFilename(filename string, dimension metadata.Dimension) string {
	ext := filepath.Ext(filename)
	filename = strings.TrimSuffix(filename, ext)
	filename = strings.TrimPrefix(filename, photosDirectory)
	return fmt.Sprintf("thumbs/%s_%dx%d%s", filename, dimension.Width, dimension.Height, jpegExtension)
}
