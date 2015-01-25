package uploadedfile

import (
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
)

type UploadedFile struct {
	filename string
	path     string
	mime     string
	hash     string
	thumbs   []*ThumbFile
}

var supportedTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/gif":  true,
	"image/png":  true,
}

func NewUploadedFile(filename, path string, thumbs []*ThumbFile) (*UploadedFile, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	buff := make([]byte, 512) // http://golang.org/pkg/net/http/#DetectContentType
	_, err = file.Read(buff)

	if err != nil {
		return nil, err
	}

	filetype := http.DetectContentType(buff)

	if _, ok := supportedTypes[filetype]; !ok {
		return nil, errors.New("Unsupported file type!")
	}

	return &UploadedFile{
		filename,
		path,
		filetype,
		"",
		thumbs,
	}, nil
}

func (this *UploadedFile) GetFilename() string {
	return this.filename
}

func (this *UploadedFile) SetFilename(filename string) {
	this.filename = filename
}

func (this *UploadedFile) GetHash() string {
	return this.hash
}

func (this *UploadedFile) SetHash(hash string) {
	this.hash = hash
}

func (this *UploadedFile) SetPath(path string) {
	// TODO: find a better location for this
	os.Remove(this.path)

	this.path = path
}

func (this *UploadedFile) GetPath() string {
	return this.path
}

func (this *UploadedFile) GetMime() string {
	return this.mime
}

func (this *UploadedFile) SetMime(mime string) {
	this.mime = mime
}

func (this *UploadedFile) GetThumbs() []*ThumbFile {
	return this.thumbs
}

func (this *UploadedFile) FileSize() (int64, error) {
	f, err := os.Open(this.path)
	if err != nil {
		return 0, err
	}

	stats, err := f.Stat()
	if err != nil {
		return 0, err
	}

	size := stats.Size()

	return size, nil
}

func (this *UploadedFile) Clean() {
	os.Remove(this.path)

	for _, thumb := range this.thumbs {
		os.Remove(thumb.path)
	}
}

func (this *UploadedFile) Dimensions() (int, int, error) {
	f, err := os.Open(this.path)
	if err != nil {
		return 0, 0, err
	}

	var cfg image.Config
	switch true {
	case this.IsGif():
		cfg, err = gif.DecodeConfig(f)
	case this.IsPng():
		cfg, err = png.DecodeConfig(f)
	case this.IsJpeg():
		cfg, err = jpeg.DecodeConfig(f)
	default:
		return 0, 0, errors.New("Invalid mime type!")
	}

	if err != nil {
		return 0, 0, err
	}

	return cfg.Width, cfg.Height, nil
}

func (this *UploadedFile) IsJpeg() bool {
	return (this.GetMime() == "image/jpeg" || this.GetMime() == "image/jpg")
}

func (this *UploadedFile) IsPng() bool {
	return this.GetMime() == "image/png"
}

func (this *UploadedFile) IsGif() bool {
	return this.GetMime() == "image/gif"
}
