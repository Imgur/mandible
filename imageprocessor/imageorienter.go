package imageprocessor

import (
	"github.com/gophergala/ImgurGo/imageprocessor/gm"
	"github.com/gophergala/ImgurGo/uploadedfile"
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
