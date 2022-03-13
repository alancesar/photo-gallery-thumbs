package usecase

import (
	"context"
	"github.com/alancesar/photo-gallery/worker/domain/metadata"
	"github.com/alancesar/photo-gallery/worker/domain/photo"
	"github.com/alancesar/photo-gallery/worker/domain/thumb"
	"io"
)

type (
	Storage interface {
		Put(ctx context.Context, img thumb.Thumbnail) error
		Get(ctx context.Context, filename string) (io.Reader, error)
	}

	Processor interface {
		FitFromReadSeeker(io.ReadSeeker, metadata.Dimension) (thumb.Image, error)
		FitAsReadSeeker(io.Reader, metadata.Dimension) (io.ReadSeeker, error)
	}

	Database interface {
		Update(ctx context.Context, id string, fields map[string]interface{}) error
	}

	Usecase interface {
		Execute(context.Context, photo.Photo) error
	}
)
