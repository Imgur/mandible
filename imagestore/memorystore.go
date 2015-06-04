package imagestore

import (
	"errors"
	"io"
	"io/ioutil"
	"strings"
)

type InMemoryImageStore struct {
	files map[string]string // name -> contents
}

func NewInMemoryImageStore() *InMemoryImageStore {
	return &InMemoryImageStore{
		files: make(map[string]string),
	}
}

func (this *InMemoryImageStore) Exists(obj *StoreObject) (bool, error) {
	_, ok := this.files[obj.Name]

	return ok, nil
}

func (this *InMemoryImageStore) Save(src io.Reader, obj *StoreObject) (*StoreObject, error) {
	data, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	this.files[obj.Name] = string(data)

	return obj, nil
}

func (this *InMemoryImageStore) Get(obj *StoreObject) (io.Reader, error) {
	data, ok := this.files[obj.Name]

	if !ok {
		return nil, errors.New("File doesn't exist")
	}

	return strings.NewReader(data), nil
}
