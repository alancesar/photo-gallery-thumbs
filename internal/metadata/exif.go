package metadata

import (
	"errors"
	"github.com/alancesar/photo-gallery/thumbs/domain/photo"
	"github.com/dsoprea/go-exif/v3"
)

func ExtractExif(bytes []byte) (photo.Exif, error) {
	rawExif, err := exif.SearchAndExtractExif(bytes)
	if err != nil {
		if errors.Is(err, exif.ErrNoExif) {
			return nil, nil
		}

		return nil, err
	}

	entries, _, err := exif.GetFlatExifData(rawExif, nil)
	if err != nil {
		return nil, err
	}

	ex := photo.Exif{}
	for _, entry := range entries {
		ex.SetTag(entry.IfdPath, entry.TagName, photo.Tag{
			ID:       entry.TagId,
			TypeName: entry.TagTypeName,
			Count:    entry.UnitCount,
			Value:    entry.Formatted,
		})
	}

	return ex, nil
}
