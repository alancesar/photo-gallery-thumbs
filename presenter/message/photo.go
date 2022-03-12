package message

import (
	"github.com/alancesar/photo-gallery/thumbs/domain/photo"
)

type Photo struct {
	ID     string        `json:"id"`
	Thumbs []photo.Image `json:"thumbs"`
}
