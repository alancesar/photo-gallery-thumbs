package config

import (
	"github.com/alancesar/photo-gallery/thumbs/domain/image"
	"gopkg.in/yaml.v2"
	"os"
)

type Thumbs struct {
	Dimensions []image.Dimension `yaml:"dimensions"`
}

type Config struct {
	Thumbs Thumbs `yaml:"thumbs"`
}

func Load(configFilepath string) (Config, error) {
	file, err := os.ReadFile(configFilepath)
	if err != nil {
		return Config{}, err
	}

	config := Config{}
	err = yaml.Unmarshal(file, &config)
	return config, err
}
