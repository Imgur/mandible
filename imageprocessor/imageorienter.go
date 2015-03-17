package imageprocessor

import (
	"github.com/Imgur/mandible/imageprocessor/gm"
	"github.com/Imgur/mandible/uploadedfile"
)

type ImageOrienter struct{}

func (this *ImageOrienter) Process(image *uploadedfile.UploadedFile) error {
	filename, err := gm.FixOrientation(image.GetPath())
	if err != nil {
		return err
	}

	image.SetPath(filename)

	return nil
}
