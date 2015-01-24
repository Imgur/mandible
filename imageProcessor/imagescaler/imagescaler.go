package imagescaler

import (
    "log"
    "fmt"
    "os"
    "io/ioutil"
    "github.com/gophergala/imgurgo/uploadedimage"
    "github.com/gophergala/imgurgo/imageprocessor/gm"
    "errors"
)

type ImageScaler struct {
    targetSize int
}

func Factory(targetSize int) (*ImageScaler) {
    return &ImageScaler{targetSize}
}

func (this *ImageScaler) Process(image *UploadedImage) error {
    switch image.GetMime() {
    case "image/jpeg":
        return this.scaleJpeg(image)
    case "image/png":
        return this.scalePng(image)
    case "image/gif":
        return this.scaleGif(image)
    }

    return errors.New("Unsuported filetype")
}

func (this *ImageScaler) scalePng(image *UploadedImage) error {
    filename, err := gm.ConvertToJpeg(image.GetFilename(), 90)
    if err != nil {
        return err
    }

    image.SetFilename(filename)
    return this.scale_jpg(image)
}

func (this *ImageScaler) scaleJpeg(image *UploadedImage) error {
    filename, err := gm.Quality(image.GetFilename(), 90)
    if err != nil {
        return err
    }

    image.SetFilename(filename)
    size, err := image.FileSize()
    if(size < this.targetSize) {
        return nil
    }

    filename, err := gm.Quality(image.GetFilename(), 70)
    if err != nil {
        return err
    }

    image.SetFilename(filename)
    size, err := image.FileSize()
    if(size < this.targetSize) {
        return nil
    }

    percent := 90
    if((size - this.targetSize) >= (15*1024*1024)) {
        percent = 30
    } else if((size - this.targetSize) >= (10*1024*1024)) {
        percent = 40
    } else if((size - this.targetSize) >= (5*1024*1024)) {
        percent = 60
    }

    for {
        gm.ResizePercent(image.GetFilename(), percent)

        image.SetFilename(filename)
        size, err := image.FileSize()
        if size == 0 || percent < 10 {
            return errors.New("Could not scale image to desired filesize")
        } else if(size < this.targetSize) {
            return nil
        }

        percent -= 10
    }
}

func (this *ImageScaler) scaleGif(image *UploadedImage) error {
    return errors.New("Unimplimented")
}