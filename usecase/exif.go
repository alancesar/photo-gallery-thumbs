package usecase

import (
	"context"
	"github.com/alancesar/photo-gallery/thumbs/domain/photo"
	"io"
)

const (
	chunkSize = 1024*128 - 1
)

type (
	ExifExtractor func([]byte) (photo.Exif, error)

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

func (e Exif) Execute(_ context.Context, reader io.Reader) (photo.Exif, error) {
	chunk := make([]byte, chunkSize)
	if _, err := reader.Read(chunk); err != nil {
		return nil, err
	}

	return e.extractor(chunk)
}
