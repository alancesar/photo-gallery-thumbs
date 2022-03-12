package thumb

import (
	"github.com/alancesar/photo-gallery/worker/domain/metadata"
	"io"
)

type (
	Image struct {
		Reader io.Reader `json:"-" firestore:"-"`
		metadata.Metadata
	}

	Thumbnail struct {
		Filename string `json:"filename" firestore:"filename"`
		Image
	}
)
