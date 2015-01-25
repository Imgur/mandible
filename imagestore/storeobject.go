package imagestore

type StoreObject struct {
	Name     string // Unique identifier
	MimeType string // i.e. image/jpg
	Type     string // i.e. thumb
	Url      string // if publicly available
}
