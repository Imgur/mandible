package imageprocessor

import (
	"github.com/Imgur/mandible/imageprocessor/processorcommand"
	"github.com/Imgur/mandible/uploadedfile"

	"log"
)

type OCRRunner struct {
	Command processorcommand.OCRCommand
}

func (this *OCRRunner) Process(image *uploadedfile.UploadedFile) error {
	result, err := this.Command.Run(image.GetPath())
	if err != nil {
		log.Printf("Error running OCR: %s", err.Error())
		return err
	}

	image.SetOCRText(result.Text)

	return nil
}

func (this *OCRRunner) String() string {
	return "OCR runner"
}

var DuelOCRStratagy = func() *OCRRunner {
	multi := processorcommand.MultiOCRCommand{}
	multi = append(multi, processorcommand.NewMemeOCR())
	multi = append(multi, processorcommand.NewStandardOCR())

	return &OCRRunner{multi}
}

var StandardOCRStratagy = func() *OCRRunner {
	return &OCRRunner{processorcommand.NewStandardOCR()}
}

var MemeOCRStratagy = func() *OCRRunner {
	return &OCRRunner{processorcommand.NewMemeOCR()}
}
