package thumbs

import (
	"github.com/alancesar/photo-gallery/thumbs/domain/image"
)

type Thumbs struct {
	Filename string        `json:"filename"`
	Images   []image.Image `json:"images"`
}
