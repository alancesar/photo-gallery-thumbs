package image

import (
	"io"
)

type Dimension struct {
	Width  int `yaml:"width" json:"width"`
	Height int `yaml:"height" json:"height"`
}

type Type string

const (
	Thumb Type = "THUMB"
)

type Metadata struct {
	ContentType string    `json:"content_type"`
	ETag        string    `json:"etag,omitempty"`
	Dimension   Dimension `json:"dimension,omitempty"`
}

type Image struct {
	Reader   io.Reader `json:"-"`
	Filename string    `json:"filename"`
	Type     Type      `json:"type,omitempty"`
	Metadata
}
