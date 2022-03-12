package thumb

import (
	"bytes"
	"github.com/alancesar/photo-gallery/worker/domain/metadata"
	"github.com/alancesar/photo-gallery/worker/pkg"
	"github.com/disintegration/imaging"
	"image"
	"io"
)

type Processor struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

func (p Processor) FitFromReadSeeker(seeker io.ReadSeeker, dimension metadata.Dimension) (Image, error) {
	_, err := seeker.Seek(0, 0)
	if err != nil {
		return Image{}, err
	}

	resized, err := fit(seeker, dimension)
	if err != nil {
		return Image{}, err
	}

	buffer, err := encode(resized)
	if err != nil {
		return Image{}, err
	}

	return Image{
		Reader: buffer,
		Metadata: metadata.Metadata{
			ContentType: pkg.JpegContentType,
			Dimension:   getDimensionFromRectangle(resized.Bounds()),
		},
	}, nil
}

func (p Processor) FitAsReadSeeker(reader io.Reader, dimension metadata.Dimension) (io.ReadSeeker, error) {
	encoded, err := fitAndEncode(reader, dimension)
	if err != nil {
		return nil, err
	}

	content, err := io.ReadAll(encoded)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(content), err
}

func fitAndEncode(reader io.Reader, dimension metadata.Dimension) (io.Reader, error) {
	resized, err := fit(reader, dimension)
	if err != nil {
		return nil, err
	}

	return encode(resized)
}

func fit(reader io.Reader, dimension metadata.Dimension) (*image.NRGBA, error) {
	decoded, err := imaging.Decode(reader)
	if err != nil {
		return nil, err
	}

	if err := validateDimensions(decoded.Bounds(), dimension); err != nil {
		return nil, err
	}

	return imaging.Fit(decoded, dimension.Width, dimension.Height, imaging.Lanczos), nil
}

func encode(resized *image.NRGBA) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	if err := imaging.Encode(buffer, resized, imaging.JPEG); err != nil {
		return nil, err
	}

	return buffer, nil
}

func validateDimensions(bounds image.Rectangle, dimension metadata.Dimension) error {
	if dimension.Width > bounds.Dx() && dimension.Height > bounds.Dy() {
		return pkg.ErrInvalidThumbSize
	}

	return nil
}

func getDimensionFromRectangle(rectangle image.Rectangle) metadata.Dimension {
	return metadata.Dimension{
		Width:  rectangle.Dx(),
		Height: rectangle.Dy(),
	}
}
