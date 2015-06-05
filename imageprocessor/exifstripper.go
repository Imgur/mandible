package imageprocessor

import (
	"github.com/Imgur/mandible/imageprocessor/processorcommand"
	"github.com/Imgur/mandible/uploadedfile"
)

type ExifStripper struct{}

func (this *ExifStripper) Process(image *uploadedfile.UploadedFile) error {
	if !image.IsJpeg() {
		return nil
	}

	err := processorcommand.StripMetadata(image.GetPath())
	if err != nil {
		return err
	}

	return nil
}

func (this *ExifStripper) String() string {
	return "EXIF stripper"
}
