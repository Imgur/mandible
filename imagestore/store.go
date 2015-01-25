package imagestore

type ImageStore interface {
	Save(src string, obj *StoreObject) error
	Exists(obj *StoreObject) (bool, error)
}
