package image

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"io"
)

type Processor struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

func (Processor) Fit(reader io.Reader, dimension Dimension) (io.Reader, Dimension, error) {
	decoded, err := imaging.Decode(reader)
	if err != nil {
		return nil, Dimension{}, err
	}

	if err := validateDimensions(decoded.Bounds(), dimension); err != nil {
		return nil, Dimension{}, err
	}

	resized := imaging.Fit(decoded, dimension.Width, dimension.Height, imaging.Lanczos)
	buffer := new(bytes.Buffer)
	if err := imaging.Encode(buffer, resized, imaging.JPEG); err != nil {
		return nil, Dimension{}, err
	}

	return buffer, getDimensionFromRectangle(resized.Bounds()), nil
}

func getDimensionFromRectangle(rectangle image.Rectangle) Dimension {
	return Dimension{
		Width:  rectangle.Dx(),
		Height: rectangle.Dy(),
	}
}

func validateDimensions(bounds image.Rectangle, dimension Dimension) error {
	if dimension.Width > bounds.Dx() || dimension.Height > bounds.Dy() {
		return fmt.Errorf("thumb size is larger then original image")
	}

	return nil
}
