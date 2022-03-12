package usecase

import (
	"context"
	"github.com/alancesar/photo-gallery/thumbs/domain/metadata"
)

const (
	chunkSize = 1024*128 - 1
)

type (
	ExifExtractor func([]byte) (metadata.Exif, error)

	Exif struct {
		storage   Storage
		extractor ExifExtractor
	}
)

func NewExif(storage Storage, extractor ExifExtractor) *Exif {
	return &Exif{
		storage:   storage,
		extractor: extractor,
	}
}

func (e Exif) Execute(ctx context.Context, filename string) (metadata.Exif, error) {
	reader, err := e.storage.Get(ctx, filename)
	if err != nil {
		return nil, err
	}

	chunk := make([]byte, chunkSize)
	if _, err := reader.Read(chunk); err != nil {
		return nil, err
	}

	return e.extractor(chunk)
}
