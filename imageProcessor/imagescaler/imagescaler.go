package imagescaler

import (
    "log"
    "fmt"
    "os"
    "io/ioutil"
)

type ImageScaler struct {
    Command OCRCommand
}

func (this *ImageScaler) Process(image string) error {
    result, err := this.Command.run(image)
    if err != nil {
        log.Printf("Error running OCR: %d", err)
        return err
    }

    log.Printf("Selected Algorithm: %s", result.Type)
    log.Printf("Final txt: %s", result.Text)

    return nil
}