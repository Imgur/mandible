package imagestore

type Factory struct {
	config string
}

func (this *Factory) NewS3() *S3ImageStore {
	return &S3ImageStore{
		bucketName: "gophergala",
		storeRoot:  "original",
	}
}

func (this *Factory) NewLocal() *LocalImageStore {
	return &LocalImageStore{
		storeRoot: "/Users/jarvis/imagestore",
	}
}

func (this *Factory) NewStoreObject(name string, mime string) *StoreObject {
	return &StoreObject{
		Path:     name,
		MimeType: mime,
	}
}
