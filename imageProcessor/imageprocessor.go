package imageprocessor

import (
	"crypto/rand"
	"log"
	"github.com/gophergala/imgurgo/uploadedfile"
	"github.com/gophergala/imgurgo/imageprocessor/imagescaler"
)

const MAX_SIZE = 1024 * 50

func init() {
	hashGetter := make(chan string)
	length := 7

	go func() {
		for {
			str := ""

			for len(str) < length {
				c := 10
				bArr := make([]byte, c)
				_, err := rand.Read(bArr)
				if err != nil {
					log.Println("error:", err)
					break
				}

				for _, b := range bArr {
					if len(str) == length {
						break
					}

					/**
					 * Each byte will be in [0, 256), but we only care about:
					 *
					 * [48, 57]     0-9
					 * [65, 90]     A-Z
					 * [97, 122]    a-z
					 *
					 * Which means that the highest bit will always be zero, since the last byte with high bit
					 * zero is 01111111 = 127 which is higher than 122. Lower our odds of having to re-roll a byte by
					 * dividing by two (right bit shift of 1).
					 */

					b = b >> 1

					// The byte is any of        0-9                  A-Z                      a-z
					byteIsAllowable := (b >= 48 && b <= 57) || (b >= 65 && b <= 90) || (b >= 97 && b <= 122)

					if byteIsAllowable {
						str += string(b)
					}
				}

			}

			hashGetter <- str
		}
	}()
}

type multiProcessType []ProcessType

func (this multiProcessType) Process(image *uploadedfile.UploadedFile) (error) {
	for _, processor := range this {
		err := processor.Process(image)
		if err != nil {
			return err
		}
	}

	return nil
}

type asyncProcessType []ProcessType

func (this asyncProcessType) Process(image *uploadedfile.UploadedFile) (error) {
	results  := make(chan bool, len(this))
	errs     := make(chan error, len(this))

	for _, processor := range this {
		go func(p ProcessType) {
			err := processor.Process(image)
			if err != nil {
				errs <- err
			}

			results <- true
		}(processor)
	}

	for i := 0; i < len(this); i++ {
		select {
		case <-results:
		case err := <-errs:
			return err
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

func (this *ImageProcessor) Run(image *uploadedfile.UploadedFile) (error) {
	return this.processor.Process(image)
}

func Factory(file *uploadedfile.UploadedFile) (*ImageProcessor, error) {
	size, err := file.FileSize()
	if err != nil {
		return &ImageProcessor{}, err
	}

	processor := multiProcessType{}

	if size > MAX_SIZE {
		processor = append(processor, imagescaler.Factory(MAX_SIZE))
	}

	return &ImageProcessor{processor}, nil
}
