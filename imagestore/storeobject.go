package imagestore

type StoreObject struct {
	Id     string // Unique identifier
	MimeType string // i.e. image/jpg
	Size     string // i.e. thumb
	Url      string // if publicly available
}
