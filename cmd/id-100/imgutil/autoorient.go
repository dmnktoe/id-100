package imgutil

import (
	"image"
	"io"

	"github.com/disintegration/imaging"
)

// DecodeAutoOriented decodes an image from r and applies EXIF-based auto-orientation.
// It returns the decoded image or an error.
func DecodeAutoOriented(r io.Reader) (image.Image, error) {
	img, err := imaging.Decode(r, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}
	return img, nil
}
