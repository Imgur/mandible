package uploadedfile

import (
	"os"
)

type UploadedFile struct {
	filename string
	path     string
	mime     string
}

func NewUploadedFile(filename, path, mime string) *UploadedFile {
	return &UploadedFile{
		filename,
		path,
		mime,
	}
}

func (this *UploadedFile) GetFilename() string {
	return this.filename
}

func (this *UploadedFile) SetFilename(filename string) {
	this.filename = filename
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
