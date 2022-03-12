package usecase

import (
	"context"
	"github.com/alancesar/photo-gallery/thumbs/domain/image"
	"github.com/alancesar/photo-gallery/thumbs/presenter/message"
	"io"
)

type Storage interface {
	Put(ctx context.Context, img image.Image) error
	Get(ctx context.Context, filename string) (io.Reader, error)
}

type Processor interface {
	Fit(reader io.Reader, dimension image.Dimension) (io.Reader, image.Dimension, error)
}

type Publisher interface {
	Publish(ctx context.Context, photo message.Photo)
}
