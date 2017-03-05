package storage

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

const (
	contentTypeHeader = "Content-Type"
)

// fileMock is mock for file operations
type fileMock interface {
	Open(name string) (*os.File, error)
	Stat(name string) (os.FileInfo, error)
}

// fileErrorMock is simple mock that returns specified errors on
// function call.
type fileErrorMock struct {
	errOpen error
	errStat error
}

func (f fileErrorMock) Open(name string) (*os.File, error) {
	return nil, f.errOpen
}

func (f fileErrorMock) Stat(name string) (os.FileInfo, error) {
	return nil, f.errStat
}

func (c *Client) fileOpen(name string) (*os.File, error) {
	if c.file != nil {
		return c.file.Open(name)
	}
	return os.Open(name)
}

func (c *Client) fileSetMockError(errOpen, errStat error) {
	c.file = &fileErrorMock{errOpen, errStat}
}

func (c *Client) fileStat(name string) (os.FileInfo, error) {
	if c.file != nil {
		return c.file.Stat(name)
	}
	return os.Stat(name)
}

// UploadFile to container
func (c *Client) UploadFile(filename, container string) error {
	f, err := c.fileOpen(filename)
	if err != nil {
		return err
	}
	stats, err := c.fileStat(filename)
	if err != nil {
		return err
	}
	ext := filepath.Ext(filename)
	mimetype := mime.TypeByExtension(ext)
	return c.Upload(f, container, stats.Name(), mimetype)
}

func (c *Client) upload(reader io.Reader, container, filename, contentType string, check bool) error {
	var etag string
	closer, ok := reader.(io.ReadCloser)
	if ok {
		defer closer.Close()
	}

	if check {
		f, err := ioutil.TempFile(os.TempDir(), filename)
		if err != nil {
			return err
		}
		stat, _ := f.Stat()
		path := stat.Name()
		hasher := md5.New()
		writer := io.MultiWriter(f, hasher)
		_, err = io.Copy(writer, reader)
		f.Close()
		if err != nil {
			return err
		}
		etag = hex.EncodeToString(hasher.Sum(nil))
		reader, err = os.Open(filepath.Join(os.TempDir(), path))
		defer os.Remove(path)
		if err != nil {
			return err
		}
	}

	request, err := c.NewRequest(putMethod, reader, container, filename)
	if err != nil {
		return err
	}
	if !blank(contentType) {
		request.Header.Add(contentTypeHeader, contentType)
	}

	if !blank(etag) {
		request.Header.Add(etagHeader, etag)
	}

	res, err := c.do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return ErrorBadResponce
	}

	return nil
}

// Upload reads all data from reader and uploads to contaier with filename and content type
func (c *Client) Upload(reader io.Reader, container, filename, contentType string) error {
	return c.upload(reader, container, filename, contentType, true)
}
