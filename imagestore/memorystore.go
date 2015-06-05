package imagestore

import (
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"sync"
)

type InMemoryImageStore struct {
	files map[string]string // name -> contents
	rw    sync.Mutex
}

func NewInMemoryImageStore() *InMemoryImageStore {
	return &InMemoryImageStore{
		files: make(map[string]string),
		rw:    sync.Mutex{},
	}
}

func (this *InMemoryImageStore) Exists(obj *StoreObject) (bool, error) {
	this.rw.Lock()

	_, ok := this.files[obj.Name]

	this.rw.Unlock()

	return ok, nil
}

func (this *InMemoryImageStore) Save(src io.Reader, obj *StoreObject) (*StoreObject, error) {
	data, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	this.rw.Lock()
	this.files[obj.Name] = string(data)
	this.rw.Unlock()

	return obj, nil
}

func (this *InMemoryImageStore) Get(obj *StoreObject) (io.Reader, error) {
	this.rw.Lock()
	data, ok := this.files[obj.Name]
	this.rw.Unlock()

	if !ok {
		return nil, errors.New("File doesn't exist")
	}

	return strings.NewReader(data), nil
}
