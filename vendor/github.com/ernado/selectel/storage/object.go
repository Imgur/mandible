package storage

import (
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	etagHeader            = "etag"
	contentLengthHeader   = "Content-Length"
	lastModifiedLayout    = time.RFC1123
	lastModifiedHeader    = "last-modified"
	objectDownloadsHeader = "X-Object-Downloads"
)

// ObjectInfo represents object info
type ObjectInfo struct {
	Size            uint64    `json:"bytes"`
	ContentType     string    `json:"content_type"`
	Downloaded      uint64    `json:"downloaded"`
	Hash            string    `json:"hash"`
	LastModifiedStr string    `json:"last_modified"`
	LastModified    time.Time `json:"-"`
	Name            string    `json:"name"`
}

type Object struct {
	name      string
	container ContainerAPI
	api       API
}

type ObjectAPI interface {
	Info() (ObjectInfo, error)
	Remove() error
	Download() ([]byte, error)
	Upload(reader io.Reader, contentType string) error
	UploadFile(filename string) error
	GetReader() (io.ReadCloser, error)
}

func (c *Client) ObjectInfo(container, filename string) (f ObjectInfo, err error) {
	request, err := c.NewRequest(headMethod, nil, container, filename)
	if err != nil {
		return f, err
	}
	res, err := c.do(request)
	if err != nil {
		return f, err
	}
	if res.StatusCode == http.StatusNotFound {
		return f, ErrorObjectNotFound
	}
	if res.StatusCode != http.StatusOK {
		return f, ErrorBadResponce
	}
	parse := func(key string) uint64 {
		v, _ := strconv.ParseUint(res.Header.Get(key), uint64Base, uint64BitSize)
		return v
	}
	f.Size = uint64(res.ContentLength)
	f.Hash = res.Header.Get(etagHeader)
	f.ContentType = res.Header.Get(contentTypeHeader)
	f.LastModified, err = time.Parse(lastModifiedLayout, res.Header.Get(lastModifiedHeader))
	f.Name = filename
	if err != nil {
		return
	}
	f.Downloaded = parse(objectDownloadsHeader)
	return
}

func (o *Object) Info() (info ObjectInfo, err error) {
	return o.container.ObjectInfo(o.name)
}

func (o *Object) Upload(reader io.Reader, contentType string) error {
	return o.container.Upload(reader, o.name, contentType)
}

func (o *Object) UploadFile(filename string) error {
	return o.container.UploadFile(filename)
}

func (o *Object) Download() ([]byte, error) {
	reader, err := o.GetReader()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}

func (o *Object) GetReader() (io.ReadCloser, error) {
	request, _ := http.NewRequest(getMethod, o.container.URL(o.name), nil)
	res, err := o.api.Do(request)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, ErrorObjectNotFound
	}
	if res.StatusCode != http.StatusOK {
		return nil, ErrorBadResponce
	}
	return res.Body, nil
}

func (o *Object) Remove() error {
	return o.container.RemoveObject(o.name)
}
