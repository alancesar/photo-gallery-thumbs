package photo

import (
	"github.com/alancesar/photo-gallery/thumbs/domain/image"
)

type Photo struct {
	Filename string `json:"filename"`
	image.Metadata
}
