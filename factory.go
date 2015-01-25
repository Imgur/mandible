package main

import (
	"github.com/gophergala/ImgurGo/imagestore"
)

type Factory struct {
	config *Configuration
}

func (this *Factory) NewS3() *imagestore.S3ImageStore {
	return imagestore.NewS3ImageStore(
		this.config.Store.S3.BucketName,
		this.config.Store.S3.StoreRoot,
		this.config.Store.S3.Region,
	)
}

func (this *Factory) NewLocal() *imagestore.LocalImageStore {
	return imagestore.NewLocalImageStore(this.config.Store.Local.StoreRoot)
}

func (this *Factory) NewStoreObject(name string, mime string, imgType string) *imagestore.StoreObject {
	return &imagestore.StoreObject{
		Name:     name,
		MimeType: mime,
		Type:     imgType,
	}
}
