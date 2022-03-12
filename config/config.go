package config

import (
	"github.com/alancesar/photo-gallery/thumbs/domain/metadata"
	"gopkg.in/yaml.v2"
	"io"
)

type Thumbs struct {
	Dimensions []metadata.Dimension `yaml:"dimensions"`
}

type Config struct {
	Thumbs Thumbs `yaml:"thumbs"`
}

func Load(reader io.Reader) (Config, error) {
	config := Config{}
	err := yaml.NewDecoder(reader).Decode(&config)
	return config, err
}
