package imagestore

import (
	"io"
	"os"
	"path"
)

// A LocalImageStore stores images on the local disk.
type LocalImageStore struct {
	storeRoot      string
	namePathMapper *NamePathMapper
}

func NewLocalImageStore(root string, mapper *NamePathMapper) *LocalImageStore {
	return &LocalImageStore{
		storeRoot:      root,
		namePathMapper: mapper,
	}
}

func (this *LocalImageStore) Exists(obj *StoreObject) (bool, error) {
	if _, err := os.Stat(this.toPath(obj)); os.IsNotExist(err) {
		return false, err
	}

	return true, nil
}

func (this *LocalImageStore) Save(src io.Reader, obj *StoreObject) (*StoreObject, error) {
	// open output file
	this.createParent(obj)
	fo, err := os.Create(this.toPath(obj))
	if err != nil {
		return nil, err
	}

	defer fo.Close()

	_, err = io.Copy(fo, src)
	if err != nil {
		return nil, err
	}

	obj.Url = this.toPath(obj)
	return obj, nil
}

func (this *LocalImageStore) Get(obj *StoreObject) (io.Reader, error) {
	reader, err := os.Open(this.toPath(obj)); 
	if err != nil {
	    return nil, err
	}

	return reader, nil
}

func (this *LocalImageStore) createParent(obj *StoreObject) {
	path := path.Dir(this.toPath(obj))

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0777)
	}
}

func (this *LocalImageStore) toPath(obj *StoreObject) string {
	return this.storeRoot + "/" + this.namePathMapper.mapToPath(obj)
}
