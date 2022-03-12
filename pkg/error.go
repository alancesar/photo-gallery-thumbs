package pkg

import (
	"errors"
)

var ErrInvalidThumbSize = errors.New("thumb size is larger than original image")
