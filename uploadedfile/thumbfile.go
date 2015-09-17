package uploadedfile

import (
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"

	"github.com/Imgur/mandible/imageprocessor/processorcommand"
	"github.com/Imgur/mandible/imageprocessor/thumbType"
)

var (
	defaultQuality = 83
)

type ThumbFile struct {
	Name        string
	Width       int
	MaxWidth    int
	Height      int
	MaxHeight   int
	Shape       string
	CropGravity string
	CropWidth   int
	CropHeight  int
	CropRatio   string
	Quality     int
	Format      string
	LocalPath   string
	StoreURI    string
}

func NewThumbFile(width, maxWidth, height, maxHeight int, name, shape, path, cropGravity string, cropWidth, cropHeight int, cropRatio string, quality int) *ThumbFile {
	if quality == 0 {
		quality = defaultQuality
	}

	return &ThumbFile{
		Name:        name,
		Width:       width,
		MaxWidth:    maxWidth,
		Height:      height,
		MaxHeight:   maxHeight,
		Shape:       shape,
		CropGravity: cropGravity,
		CropWidth:   cropWidth,
		CropHeight:  cropHeight,
		CropRatio:   cropRatio,
		Quality:     quality,
		Format:      "",
		LocalPath:   path,
		StoreURI:    "",
	}
}

func (this *ThumbFile) SetPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Error when creating thumbnail %s", this.Name))
	}

	this.LocalPath = path

	return nil
}

func (this *ThumbFile) GetPath() string {
	return this.LocalPath
}

func (this *ThumbFile) ComputeWidth(original *UploadedFile) int {
	width := this.Width

	oWidth, _, err := original.Dimensions()
	if err != nil {
		return 0
	}

	if this.MaxWidth > 0 {
		width = int(math.Min(float64(oWidth), float64(this.MaxWidth)))
	}

	return width
}

func (this *ThumbFile) ComputeHeight(original *UploadedFile) int {
	height := this.Height

	_, oHeight, err := original.Dimensions()
	if err != nil {
		return 0
	}

	if this.MaxHeight > 0 {
		height = int(math.Min(float64(oHeight), float64(this.MaxHeight)))
	}

	return height
}

func (this *ThumbFile) ComputeCrop(original *UploadedFile) (int, int, error) {
	re := regexp.MustCompile("(.*):(.*)")
	matches := re.FindStringSubmatch(this.CropRatio)
	if len(matches) != 3 {
		return 0, 0, errors.New("Invalid crop_ratio")
	}

	wRatio, werr := strconv.ParseFloat(matches[1], 64)
	hRatio, herr := strconv.ParseFloat(matches[2], 64)
	if werr != nil || herr != nil {
		return 0, 0, errors.New("Invalid crop_ratio")
	}

	var cropWidth, cropHeight float64

	if wRatio >= hRatio {
		wRatio = wRatio / hRatio
		hRatio = 1
		cropWidth = math.Ceil(float64(this.ComputeHeight(original)) * wRatio)
		cropHeight = math.Ceil(float64(this.ComputeHeight(original)) * hRatio)
	} else {
		hRatio = hRatio / wRatio
		wRatio = 1
		cropWidth = math.Ceil(float64(this.ComputeWidth(original)) * wRatio)
		cropHeight = math.Ceil(float64(this.ComputeWidth(original)) * hRatio)
	}

	return int(cropWidth), int(cropHeight), nil
}

func (this *ThumbFile) Process(original *UploadedFile) error {
	switch this.Shape {
	case "circle":
		return this.processCircle(original)
	case "thumb":
		return this.processThumb(original)
	case "square":
		return this.processSquare(original)
	case "custom":
		return this.processCustom(original)

	}

	return errors.New("Invalid thumb shape " + this.Shape)
}

func (this *ThumbFile) String() string {
	return fmt.Sprintf("Thumbnail of <%s>", this.Name)
}

func (this *ThumbFile) processSquare(original *UploadedFile) error {
	filename, err := processorcommand.SquareThumb(original.GetPath(), this.Name, this.Width, thumbType.FromMime(original.GetMime()))
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}

func (this *ThumbFile) processCircle(original *UploadedFile) error {
	filename, err := processorcommand.CircleThumb(original.GetPath(), this.Name, this.Width, thumbType.FromMime(original.GetMime()))
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}

func (this *ThumbFile) processThumb(original *UploadedFile) error {
	filename, err := processorcommand.Thumb(original.GetPath(), this.Name, this.Width, this.Height, thumbType.FromMime(original.GetMime()))
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}

func (this *ThumbFile) processCustom(original *UploadedFile) error {
	cropWidth := this.CropWidth
	cropHeight := this.CropHeight
	var err error

	if this.CropRatio != "" {
		cropWidth, cropHeight, err = this.ComputeCrop(original)
		if err != nil {
			return err
		}
	}

	filename, err := processorcommand.CustomThumb(original.GetPath(), this.Name, this.ComputeWidth(original), this.ComputeHeight(original), this.CropGravity, cropWidth, cropHeight, this.Quality, thumbType.FromMime(original.GetMime()))
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}
