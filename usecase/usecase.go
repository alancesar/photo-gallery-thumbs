package usecase

import (
	"context"
	"github.com/alancesar/photo-gallery/thumbs/domain/photo"
	"github.com/alancesar/photo-gallery/thumbs/presenter/message"
	"io"
)

type Storage interface {
	Put(ctx context.Context, img photo.Image) error
	Get(ctx context.Context, filename string) (io.Reader, error)
}

type Processor interface {
	Fit(reader io.Reader, dimension photo.Dimension) (io.Reader, photo.Dimension, error)
}

type Publisher interface {
	Publish(ctx context.Context, photo message.Photo)
}
