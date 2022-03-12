package usecase

import (
	"context"
	"github.com/alancesar/photo-gallery/worker/domain/metadata"
	"github.com/alancesar/photo-gallery/worker/domain/thumb"
	"github.com/alancesar/photo-gallery/worker/presenter/message"
	"io"
)

type Storage interface {
	Put(ctx context.Context, img thumb.Thumbnail) error
	Get(ctx context.Context, filename string) (io.Reader, error)
}

type Processor interface {
	FitFromReadSeeker(io.ReadSeeker, metadata.Dimension) (thumb.Image, error)
	FitAsReadSeeker(io.Reader, metadata.Dimension) (io.ReadSeeker, error)
}

type Publisher interface {
	Publish(ctx context.Context, photo message.Photo)
}
