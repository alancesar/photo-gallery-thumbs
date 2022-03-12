package metadata

type (
	Dimension struct {
		Width  int `yaml:"width" json:"width"`
		Height int `yaml:"height" json:"height"`
	}

	Metadata struct {
		ContentType string    `json:"content_type"`
		Dimension   Dimension `json:"dimension,omitempty"`
	}

	Tag struct {
		ID       uint16
		TypeName string
		Count    uint32
		Value    string
	}

	Path map[string]Tag
	Exif map[string]Path
)

func (e Exif) SetTag(path, name string, tag Tag) {
	if e[path] == nil {
		e[path] = Path{}
	}

	e[path][name] = tag
}

func GetLargestDimension(dimensions ...Dimension) (largest Dimension) {
	for _, dimension := range dimensions {
		if dimension.Width > largest.Width {
			largest = dimension
		}
	}

	return
}
