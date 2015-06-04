package imageprocessor

import (
	"github.com/Imgur/mandible/imageprocessor/processorcommand"
	"github.com/Imgur/mandible/uploadedfile"
)

type ImageOrienter struct{}

func (this *ImageOrienter) Process(image *uploadedfile.UploadedFile) error {
	filename, err := processorcommand.FixOrientation(image.GetPath())
	if err != nil {
		return err
	}

	image.SetPath(filename)

	return nil
}
