package imagestore

import (
	"bufio"
	"io"
	"os"
	"path"
)

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

func (this *LocalImageStore) Save(src string, obj *StoreObject) (*StoreObject, error) {
	// open input file
	fi, err := os.Open(src)
	if err != nil {
		return nil, err
	}

	defer fi.Close()

	// make a read buffer
	r := bufio.NewReader(fi)

	// open output file
	this.createParent(obj)
	fo, err := os.Create(this.toPath(obj))
	if err != nil {
		return nil, err
	}

	defer fo.Close()

	// make a write buffer
	w := bufio.NewWriter(fo)

	// make a buffer to keep chunks that are read
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}

		if n == 0 {
			break
		}

		// write a chunk
		if _, err := w.Write(buf[:n]); err != nil {
			return nil, err
		}
	}

	if err = w.Flush(); err != nil {
		return nil, err
	}

	obj.Url = this.toPath(obj)
	return obj, nil
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
