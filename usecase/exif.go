package usecase

import (
	"context"
	"fmt"
	"github.com/alancesar/photo-gallery/worker/domain/metadata"
	"github.com/alancesar/photo-gallery/worker/domain/photo"
	"io"
	"os"
	"path/filepath"
)

const (
	exifFieldName = "exif"
)

type (
	ExifTool func(filename string) (metadata.Exif, error)

	Exif struct {
		s  Storage
		e  ExifTool
		db Database
	}
)

func NewExif(storage Storage, tool ExifTool, database Database) *Exif {
	return &Exif{
		s:  storage,
		e:  tool,
		db: database,
	}
}

func (e Exif) Execute(ctx context.Context, photo photo.Photo) error {
	reader, err := e.s.Get(ctx, photo.Filename)
	if err != nil {
		return err
	}

	file, err := createTempFile(photo.Filename)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}()

	exif, err := e.e(file.Name())
	if err != nil {
		return err
	}

	return e.db.Update(ctx, photo.ID, map[string]interface{}{
		exifFieldName: exif,
	})
}

func createTempFile(filename string) (*os.File, error) {
	base := filepath.Base(filename)
	file, err := os.CreateTemp("", fmt.Sprintf("*_%s", base))
	if err != nil {
		return nil, err
	}
	return file, nil
}
