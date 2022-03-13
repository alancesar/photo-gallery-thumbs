package tool

import (
	"github.com/alancesar/photo-gallery/worker/domain/metadata"
	"github.com/barasher/go-exiftool"
)

func Exif(filename string) (metadata.Exif, error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = et.Close()
	}()

	md := et.ExtractMetadata(filename)
	if len(md) == 0 {
		return nil, nil
	}

	return md[0].Fields, nil
}
