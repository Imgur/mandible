package imagestore

type ImageStore interface {
	Save(src string, obj *StoreObject) (*StoreObject, error)
	Exists(obj *StoreObject) (bool, error)
}

type ImageStores []ImageStore

func (this *ImageStores) Save(src string, obj *StoreObject) {
	// TODO
}
