package uploadedfile

import (
	"errors"
	"fmt"
	"os"

	"github.com/gophergala/ImgurGo/imageprocessor/gm"
)

type ThumbFile struct {
	name   string
	width  int
	height int
	shape  string
	path   string
}

func NewThumbFile(width, height int, name, shape, path string) *ThumbFile {
	return &ThumbFile{
		name,
		width,
		height,
		shape,
		path,
	}
}

func (this *ThumbFile) GetName() string {
	return this.name
}

func (this *ThumbFile) SetName(name string) {
	this.name = name
}

func (this *ThumbFile) GetHeight() int {
	return this.height
}

func (this *ThumbFile) SetHeight(h int) {
	this.height = h
}

func (this *ThumbFile) GetWidth() int {
	return this.width
}

func (this *ThumbFile) SetWidth(h int) {
	this.width = h
}

func (this *ThumbFile) GetShape() string {
	return this.shape
}

func (this *ThumbFile) SetShape(shape string) {
	this.shape = shape
}

func (this *ThumbFile) GetPath() string {
	return this.path
}

func (this *ThumbFile) SetPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Error when creating thumbnail %s", this.GetName()))
	}

	this.path = path

	return nil
}

func (this *ThumbFile) Process(original *UploadedFile) error {
	switch this.shape {
	case "circle":
		return this.processCircle(original)
	case "thumb":
		return this.processThumb(original)
	case "square":
		return this.processSquare(original)

	}

	return errors.New("Invalid thumb shape " + this.shape)
}

func (this *ThumbFile) processSquare(original *UploadedFile) error {
	filename, err := gm.SquareThumb(original.GetPath(), this.GetName(), this.GetWidth())
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}

func (this *ThumbFile) processCircle(original *UploadedFile) error {
	filename, err := gm.CircleThumb(original.GetPath(), this.GetName(), this.GetWidth())
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}

func (this *ThumbFile) processThumb(original *UploadedFile) error {
	filename, err := gm.Thumb(original.GetPath(), this.GetName(), this.GetWidth(), this.GetHeight())
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}
