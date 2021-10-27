package image

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"io"
	"io/ioutil"
	"os"
)

type ImagingProcessor struct{}

func NewImagingProcessor() *ImagingProcessor {
	return &ImagingProcessor{}
}

func (ImagingProcessor) Fit(reader io.Reader, dimension Dimension) (io.Reader, Dimension, error) {
	decoded, err := imaging.Decode(reader)
	if err != nil {
		return nil, Dimension{}, err
	}

	if err := validateDimensions(decoded.Bounds(), dimension); err != nil {
		return nil, Dimension{}, err
	}

	resized := imaging.Fit(decoded, dimension.Width, dimension.Height, imaging.Lanczos)
	file, err := ioutil.TempFile("", "thumb.*.jpg")
	if err != nil {
		return nil, Dimension{}, err
	}
	defer func(f *os.File) {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}(file)

	if err := imaging.Save(resized, file.Name()); err != nil {
		return nil, Dimension{}, err
	}

	raw, err := io.ReadAll(file)
	if err != nil {
		return nil, Dimension{}, err
	}

	return io.NopCloser(bytes.NewBuffer(raw)), getDimensionFromRectangle(resized.Bounds()), nil
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
