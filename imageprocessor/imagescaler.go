package imageprocessor

import (
	"errors"

	"github.com/gophergala/ImgurGo/imageprocessor/gm"
	"github.com/gophergala/ImgurGo/uploadedfile"
)

type ImageScaler struct {
	targetSize int64
}

func (this *ImageScaler) Process(image *uploadedfile.UploadedFile) error {
	switch image.GetMime() {
	case "image/jpeg", "image/jpg":
		return this.scaleJpeg(image)
	case "image/png":
		return this.scalePng(image)
	case "image/gif":
		return this.scaleGif(image)
	}

	return errors.New("Unsuported filetype")
}

func (this *ImageScaler) scalePng(image *uploadedfile.UploadedFile) error {
	filename, err := gm.ConvertToJpeg(image.GetPath())
	if err != nil {
		return err
	}

	image.SetPath(filename)
	image.SetMime("image/jpeg")
	return this.scaleJpeg(image)
}

func (this *ImageScaler) scaleJpeg(image *uploadedfile.UploadedFile) error {
	filename, err := gm.Quality(image.GetPath(), 90)
	if err != nil {
		return err
	}

	image.SetPath(filename)
	size, err := image.FileSize()
	if size < this.targetSize {
		return nil
	}

	filename, err = gm.Quality(image.GetPath(), 70)
	if err != nil {
		return err
	}

	image.SetPath(filename)
	size, err = image.FileSize()
	if size < this.targetSize {
		return nil
	}

	percent := 90
	if (size - this.targetSize) >= (15 * 1024 * 1024) {
		percent = 30
	} else if (size - this.targetSize) >= (10 * 1024 * 1024) {
		percent = 40
	} else if (size - this.targetSize) >= (5 * 1024 * 1024) {
		percent = 60
	}

	for {
		filename, err = gm.ResizePercent(image.GetPath(), percent)
		if err != nil {
			return err
		}

		image.SetPath(filename)
		size, err := image.FileSize()
		if err != nil {
			return err
		} else if size == 0 || percent < 10 {
			return errors.New("Could not scale image to desired filesize")
		} else if size < this.targetSize {
			return nil
		}

		percent -= 10
	}
}

func (this *ImageScaler) scaleGif(image *uploadedfile.UploadedFile) error {
	return errors.New("Unimplimented")
}
