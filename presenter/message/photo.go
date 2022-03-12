package message

import (
	"github.com/alancesar/photo-gallery/thumbs/domain/metadata"
	"github.com/alancesar/photo-gallery/thumbs/domain/thumb"
)

type Photo struct {
	ID     string            `json:"id"`
	Exif   metadata.Exif     `json:"exif"`
	Thumbs []thumb.Thumbnail `json:"thumbs"`
}
