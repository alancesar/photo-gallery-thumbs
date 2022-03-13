package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/alancesar/photo-gallery/worker/domain/metadata"
	"github.com/alancesar/photo-gallery/worker/domain/photo"
	"github.com/alancesar/photo-gallery/worker/domain/thumb"
	"github.com/alancesar/photo-gallery/worker/pkg"
	"golang.org/x/sync/errgroup"
	"io"
	"path/filepath"
	"strings"
)

const (
	thumbnailsFieldName = "thumbs"
	jpegExtension       = ".jpeg"
)

type (
	ThumbnailsWorker struct {
		s  Storage
		p  Processor
		db Database
		d  []metadata.Dimension
	}
)

func NewThumbnails(storage Storage, processor Processor, database Database, dimensions []metadata.Dimension) *ThumbnailsWorker {
	return &ThumbnailsWorker{
		s:  storage,
		p:  processor,
		db: database,
		d:  dimensions,
	}
}

func (t ThumbnailsWorker) Execute(ctx context.Context, photo photo.Photo) error {
	original, err := t.s.Get(ctx, photo.Filename)
	if err != nil {
		return err
	}

	largestDimension := metadata.GetLargestDimension(t.d...)
	sample, err := t.p.FitAsReadSeeker(original, largestDimension)
	if err != nil {
		return err
	}

	thumbnails, err := t.generateThumbnails(sample, photo.Filename, t.d)
	if err != nil {
		return err
	}

	if err := t.putOnStorage(ctx, thumbnails); err != nil {
		return err
	}

	return t.db.Update(ctx, photo.ID, map[string]interface{}{
		thumbnailsFieldName: thumbnails,
	})
}

func (t ThumbnailsWorker) generateThumbnails(seeker io.ReadSeeker, filename string, dimensions []metadata.Dimension) ([]thumb.Thumbnail, error) {
	var thumbnails []thumb.Thumbnail
	for _, dimension := range dimensions {
		resized, err := t.p.FitFromReadSeeker(seeker, dimension)
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
			return t.s.Put(ctx, thumbnail)
		}

		group.Go(worker)
	}

	return group.Wait()
}

func createThumbFilename(filename string, dimension metadata.Dimension) string {
	ext := filepath.Ext(filename)
	filename = strings.TrimSuffix(filename, ext)
	filename = filepath.Base(filename)
	return fmt.Sprintf("thumbs/%s_%dx%d%s", filename, dimension.Width, dimension.Height, jpegExtension)
}
