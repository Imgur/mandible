package imagestore

import (
	"io"
)

type ImageStore interface {
	Save(src io.Reader, obj *StoreObject) (*StoreObject, error)
	Exists(obj *StoreObject) (bool, error)
    Get(obj *StoreObject) (io.Reader, error)
}

type ImageStores []ImageStore

func (this *ImageStores) Save(src string, obj *StoreObject) {
	// TODO
}

func (this *ImageStores) Exists(obj *StoreObject) (bool, error) {
	return false, nil
}
