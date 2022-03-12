package thumb

import (
	"github.com/alancesar/photo-gallery/worker/domain/metadata"
	"io"
)

type (
	Image struct {
		Reader io.Reader `json:"-"`
		metadata.Metadata
	}

	Thumbnail struct {
		Filename string `json:"filename"`
		Image
	}
)
