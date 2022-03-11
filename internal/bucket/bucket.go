package bucket

import (
	"cloud.google.com/go/storage"
	"context"
	"github.com/alancesar/photo-gallery/thumbs/domain/image"
	"io"
)

const (
	photosDirectoryName = "photos/"
)

type (
	Bucket struct {
		handle *storage.BucketHandle
	}
)

func New(handle *storage.BucketHandle) *Bucket {
	return &Bucket{
		handle: handle,
	}
}

func (s *Bucket) Put(ctx context.Context, image image.Image) error {
	writer := s.handle.Object(image.Filename).NewWriter(ctx)
	if _, err := io.Copy(writer, image.Reader); err != nil {
		return err
	}

	return writer.Close()
}

func (s *Bucket) Get(ctx context.Context, filename string) (io.Reader, error) {
	return s.handle.Object(filename).NewReader(ctx)
}
