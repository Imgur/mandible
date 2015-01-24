package imagestore

import (
	"bufio"
	"io"
	"os"
)

type LocalImageStore struct {
	storeRoot string
}

func NewLocalImageStore(root string) *LocalImageStore {
	return &LocalImageStore{
		storeRoot: root,
	}
}

func (this *LocalImageStore) Exists(obj *StoreObject) bool {
	if _, err := os.Stat(this.storeRoot + "/" + obj.Path); os.IsNotExist(err) {
		return false
	}

	return true
}

func (this *LocalImageStore) Save(src string, obj *StoreObject) error {
	// open input file
	fi, err := os.Open(src)
	if err != nil {
		panic(err)
	}

	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	// make a read buffer
	r := bufio.NewReader(fi)

	// open output file
	fo, err := os.Create(this.storeRoot + "/" + obj.Path)
	if err != nil {
		panic(err)
	}

	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	// make a write buffer
	w := bufio.NewWriter(fo)

	// make a buffer to keep chunks that are read
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}

		if n == 0 {
			break
		}

		// write a chunk
		if _, err := w.Write(buf[:n]); err != nil {
			panic(err)
		}
	}

	if err = w.Flush(); err != nil {
		panic(err)
	}

	return nil
}
