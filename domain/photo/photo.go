package photo

import (
	"github.com/alancesar/photo-gallery/thumbs/domain/image"
)

type Photo struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	image.Metadata
}
