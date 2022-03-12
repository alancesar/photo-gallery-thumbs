package photo

import (
	"io"
)

type Dimension struct {
	Width  int `yaml:"width" json:"width"`
	Height int `yaml:"height" json:"height"`
}

type Metadata struct {
	ContentType string    `json:"content_type"`
	Dimension   Dimension `json:"dimension,omitempty"`
}

type Image struct {
	Reader   io.Reader `json:"-"`
	Filename string    `json:"filename"`
	Metadata
}
