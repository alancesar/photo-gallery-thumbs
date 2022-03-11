package message

import "github.com/alancesar/photo-gallery/thumbs/domain/image"

type Photo struct {
	ID     string        `json:"id"`
	Thumbs []image.Image `json:"thumbs"`
}
