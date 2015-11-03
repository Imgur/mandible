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
	name        string
	width       int
	maxWidth    int
	height      int
	maxHeight   int
	shape       string
	cropGravity string
	cropWidth   int
	cropHeight  int
	cropRatio   string
	quality     int
	format      string
	localPath   string
	storeURI    string
}

func NewThumbFile(width, maxWidth, height, maxHeight int, name, shape, path, cropGravity string, cropWidth, cropHeight int, cropRatio string, quality int) *ThumbFile {
	if quality == 0 {
		quality = defaultQuality
	}

	return &ThumbFile{
		name,
		width,
		maxWidth,
		height,
		maxHeight,
		shape,
		cropGravity,
		cropWidth,
		cropHeight,
		cropRatio,
		quality,
		"",
		path,
		"",
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

func (this *ThumbFile) GetMaxHeight() int {
	return this.maxHeight
}

func (this *ThumbFile) SetMaxHeight(h int) {
	this.maxHeight = h
}

func (this *ThumbFile) GetWidth() int {
	return this.width
}

func (this *ThumbFile) SetWidth(h int) {
	this.width = h
}

func (this *ThumbFile) GetMaxWidth() int {
	return this.maxWidth
}

func (this *ThumbFile) SetMaxWidth(h int) {
	this.maxWidth = h
}

func (this *ThumbFile) GetShape() string {
	return this.shape
}

func (this *ThumbFile) SetShape(shape string) {
	this.shape = shape
}

func (this *ThumbFile) GetPath() string {
	return this.localPath
}

func (this *ThumbFile) SetPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Error when creating thumbnail %s", this.GetName()))
	}

	this.localPath = path

	return nil
}

func (this *ThumbFile) SetCropGravity(gravity string) {
	this.cropGravity = gravity
}

func (this *ThumbFile) GetCropGravity() string {
	return this.cropGravity
}

func (this *ThumbFile) SetCropWidth(width int) {
	this.cropWidth = width
}

func (this *ThumbFile) GetCropWidth() int {
	return this.cropWidth
}

func (this *ThumbFile) SetCropHeight(height int) {
	this.cropHeight = height
}

func (this *ThumbFile) GetCropHeight() int {
	return this.cropHeight
}

func (this *ThumbFile) SetCropRatio(ratio string) {
	this.cropRatio = ratio
}

func (this *ThumbFile) GetCropRatio() string {
	return this.cropRatio
}

func (this *ThumbFile) SetQuality(quality int) {
	this.quality = quality
}

func (this *ThumbFile) GetQuality() int {
	return this.quality
}

func (this *ThumbFile) ComputeWidth(original *UploadedFile) int {
	width := this.GetWidth()

	oWidth, _, err := original.Dimensions()
	if err != nil {
		return 0
	}

	if this.GetMaxWidth() > 0 {
		width = int(math.Min(float64(oWidth), float64(this.GetMaxWidth())))
	}

	return width
}

func (this *ThumbFile) ComputeHeight(original *UploadedFile) int {
	height := this.GetHeight()

	_, oHeight, err := original.Dimensions()
	if err != nil {
		return 0
	}

	if this.GetMaxHeight() > 0 {
		height = int(math.Min(float64(oHeight), float64(this.GetMaxHeight())))
	}

	return height
}

func (this *ThumbFile) ComputeCrop(original *UploadedFile) (int, int, error) {
	re := regexp.MustCompile("(.*):(.*)")
	matches := re.FindStringSubmatch(this.GetCropRatio())
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
	switch this.shape {
	case "circle":
		return this.processCircle(original)
	case "thumb":
		return this.processThumb(original)
	case "square":
		return this.processSquare(original)
	case "custom":
		return this.processCustom(original)

	}

	return errors.New("Invalid thumb shape " + this.shape)
}

func (this *ThumbFile) String() string {
	return fmt.Sprintf("Thumbnail of <%s>", this.name)
}

func (this *ThumbFile) processSquare(original *UploadedFile) error {
	filename, err := processorcommand.SquareThumb(original.GetPath(), this.GetName(), this.GetWidth(), thumbType.FromMime(original.GetMime()))
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}

func (this *ThumbFile) processCircle(original *UploadedFile) error {
	filename, err := processorcommand.CircleThumb(original.GetPath(), this.GetName(), this.GetWidth(), thumbType.FromMime(original.GetMime()))
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}

func (this *ThumbFile) processThumb(original *UploadedFile) error {
	filename, err := processorcommand.Thumb(original.GetPath(), this.GetName(), this.GetWidth(), this.GetHeight(), thumbType.FromMime(original.GetMime()))
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}

func (this *ThumbFile) processCustom(original *UploadedFile) error {
	cropWidth := this.GetCropWidth()
	cropHeight := this.GetCropHeight()
	var err error

	if this.GetCropRatio() != "" {
		cropWidth, cropHeight, err = this.ComputeCrop(original)
		if err != nil {
			return err
		}
	}

	filename, err := processorcommand.CustomThumb(original.GetPath(), this.GetName(), this.ComputeWidth(original), this.ComputeHeight(original), this.GetCropGravity(), cropWidth, cropHeight, this.GetQuality(), thumbType.FromMime(original.GetMime()))
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}
