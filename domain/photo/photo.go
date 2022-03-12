package photo

import "github.com/alancesar/photo-gallery/worker/domain/metadata"

type Photo struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	metadata.Metadata
}
