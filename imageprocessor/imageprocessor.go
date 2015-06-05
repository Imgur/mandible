package imageprocessor

import (
	"fmt"
	"strings"

	"github.com/Imgur/mandible/config"
	"github.com/Imgur/mandible/uploadedfile"
)

type ProcessType interface {
	Process(image *uploadedfile.UploadedFile) error
	String() string
}

type multiProcessType []ProcessType

func (this multiProcessType) Process(image *uploadedfile.UploadedFile) error {
	for _, processor := range this {
		err := processor.Process(image)
		if err != nil {
			return fmt.Errorf("Error multiprocessing on %s: %s", processor.String(), err.Error())
		}
	}

	return nil
}

func (this multiProcessType) String() string {
	processes := make([]string, 0)
	for _, p := range this {
		processes = append(processes, p.String())
	}
	return "Multiple processes <" + strings.Join(processes, ", ") + ">"
}

type asyncProcessType []ProcessType

func (this asyncProcessType) Process(image *uploadedfile.UploadedFile) error {
	errs := make(chan error, len(this))

	for _, processor := range this {
		go func(p ProcessType) {
			err := p.Process(image)
			if err != nil {
				errs <- fmt.Errorf("Error asynchronously processing on %s: %s", p.String(), err.Error())
			} else {
				errs <- nil
			}
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

func (this asyncProcessType) String() string {
	processes := make([]string, 0)
	for _, p := range this {
		processes = append(processes, p.String())
	}
	return "Async processes <" + strings.Join(processes, ", ") + ">"
}

type ImageProcessor struct {
	processor ProcessType
}

func (this *ImageProcessor) Run(image *uploadedfile.UploadedFile) error {
	return this.processor.Process(image)
}

type ImageProcessorStrategy func(*config.Configuration, *uploadedfile.UploadedFile) (*ImageProcessor, error)

// Just do nothing to the file after it's uploaded...
var PassthroughStrategy = func(cfg *config.Configuration, file *uploadedfile.UploadedFile) (*ImageProcessor, error) {
	return &ImageProcessor{multiProcessType{}}, nil
}

var EverythingStrategy = func(cfg *config.Configuration, file *uploadedfile.UploadedFile) (*ImageProcessor, error) {
	size, err := file.FileSize()
	if err != nil {
		return &ImageProcessor{}, err
	}

	processor := multiProcessType{}
	processor = append(processor, &ImageOrienter{})
	processor = append(processor, &CompressLosslessly{})
	processor = append(processor, &ExifStripper{})

	if size > cfg.MaxFileSize {
		processor = append(processor, &ImageScaler{cfg.MaxFileSize})
	}

	async := asyncProcessType{}

	async = append(async, DuelOCRStratagy())
	for _, t := range file.GetThumbs() {
		async = append(async, t)
	}

	if len(async) > 0 {
		processor = append(processor, async)
	}

	return &ImageProcessor{processor}, nil
}
