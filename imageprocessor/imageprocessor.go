package imageprocessor

import (
	"github.com/gophergala/ImgurGo/uploadedfile"
)

type multiProcessType []ProcessType

func (this multiProcessType) Process(image *uploadedfile.UploadedFile) error {
	for _, processor := range this {
		err := processor.Process(image)
		if err != nil {
			return err
		}
	}

	return nil
}

type asyncProcessType []ProcessType

func (this asyncProcessType) Process(image *uploadedfile.UploadedFile) error {
	errs := make(chan error, len(this))

	for _, processor := range this {
		go func(p ProcessType) {
			errs <- p.Process(image)
		}(processor)
	}

	for i := 0; i < len(this); i++ {
		select {
		case err := <-errs:
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type ProcessType interface {
	Process(image *uploadedfile.UploadedFile) error
}

type ImageProcessor struct {
	processor ProcessType
}

func (this *ImageProcessor) Run(image *uploadedfile.UploadedFile) error {
	return this.processor.Process(image)
}

func Factory(maxFileSize int64, file *uploadedfile.UploadedFile) (*ImageProcessor, error) {
	size, err := file.FileSize()
	if err != nil {
		return &ImageProcessor{}, err
	}

	processor := multiProcessType{}
	processor = append(processor, &ImageOrienter{})

	if size > maxFileSize {
		processor = append(processor, &ImageScaler{maxFileSize})
	}

	async := asyncProcessType{}

	for _, t := range file.GetThumbs() {
		async = append(async, t)
	}

	if len(async) > 0 {
		processor = append(processor, async)
	}

	return &ImageProcessor{processor}, nil
}
