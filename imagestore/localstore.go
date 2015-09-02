package imagestore

import (
	"fmt"
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

	fmt.Println("Saving File " + this.toPath(obj))
	size, err := io.Copy(fo, src)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	fmt.Println("Saved File " + this.toPath(obj) + " of size " + string(size))
	obj.Url = this.toPath(obj)
	return obj, nil
}

func (this *LocalImageStore) Get(obj *StoreObject) (io.ReadCloser, error) {
	reader, err := os.Open(this.toPath(obj))
	if err != nil {
		return nil, err
	}

	return reader, nil
}

func (this *LocalImageStore) String() string {
	return "LocalStore"
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
