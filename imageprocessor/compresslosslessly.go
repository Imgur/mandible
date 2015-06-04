package imageprocessor

import (
	"errors"

	"github.com/Imgur/mandible/imageprocessor/processorcommand"
	"github.com/Imgur/mandible/uploadedfile"
)

type CompressLosslessly struct{}

func (this *CompressLosslessly) Process(image *uploadedfile.UploadedFile) error {
	if image.IsJpeg() {
		return this.compressJpeg(image)
	}

	if image.IsPng() {
		return this.compressPng(image)
	}

	if image.IsGif() {
		return nil
	}

	return errors.New("Unsuported filetype")
}

func (this *CompressLosslessly) compressPng(image *uploadedfile.UploadedFile) error {
	filename, err := processorcommand.Optipng(image.GetPath())
	if err != nil {
		return err
	}

	image.SetPath(filename)

	return nil
}

func (this *CompressLosslessly) compressJpeg(image *uploadedfile.UploadedFile) error {
	filename, err := processorcommand.Jpegtran(image.GetPath())
	if err != nil {
		return err
	}

	image.SetPath(filename)

	return nil
}
