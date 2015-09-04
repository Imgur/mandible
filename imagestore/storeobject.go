package imagestore

type StorableObject interface {
	GetPath() string
}

type StoreObject struct {
	Id       string // Unique identifier
	MimeType string // i.e. image/jpg
	Size     string // i.e. thumb
	Url      string // if publicly available
}

func (this *StoreObject) Store(s StorableObject, store ImageStore) error {
	path := s.GetPath()

	obj, err := store.Save(path, this)
	if err != nil {
		return err
	}

	this.Url = obj.Url

	return nil
}
