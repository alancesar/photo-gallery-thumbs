package metadata

type (
	Dimension struct {
		Width  int `yaml:"width" json:"width" firestore:"width"`
		Height int `yaml:"height" json:"height" firestore:"height"`
	}

	Metadata struct {
		ContentType string    `json:"content_type" firestore:"content_type"`
		Dimension   Dimension `json:"dimension,omitempty" firestore:"dimension,omitempty"`
	}

	Exif map[string]interface{}
)

func GetLargestDimension(dimensions ...Dimension) (largest Dimension) {
	for _, dimension := range dimensions {
		if dimension.Width > largest.Width {
			largest = dimension
		}
	}

	return
}
