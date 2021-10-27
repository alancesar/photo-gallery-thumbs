package storage

import (
	"context"
	"github.com/alancesar/photo-gallery/thumbs/image"
	"github.com/minio/minio-go/v7"
	"io"
)

type Storage struct {
	client *minio.Client
	bucket string
}

func NewMinioStorage(client *minio.Client, bucket string) *Storage {
	return &Storage{
		client: client,
		bucket: bucket,
	}
}

func (s *Storage) Put(ctx context.Context, image image.Image) error {
	_, err := s.client.PutObject(ctx, s.bucket, image.Filename, image.Reader, -1, minio.PutObjectOptions{
		ContentType: image.ContentType,
	})

	return err
}

func (s *Storage) Get(ctx context.Context, filename string) (io.ReadSeeker, error) {
	return s.client.GetObject(ctx, s.bucket, filename, minio.GetObjectOptions{})
}
