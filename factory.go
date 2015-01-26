package main

import (
	"github.com/gophergala/ImgurGo/imagestore"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"log"
	"os"
)

type Factory struct {
	config *Configuration
}

func (this *Factory) NewImageStores() []imagestore.ImageStore {
	stores := []imagestore.ImageStore{}

	for _, configWrapper := range this.config.Stores {
		switch configWrapper["Type"] {
		case "s3":
			store := this.NewS3ImageStore(configWrapper)
			stores = append(stores, store)
		case "local":
			store := this.NewLocalImageStore(configWrapper)
			stores = append(stores, store)
		default:
			log.Fatal("Unsupported store %s", configWrapper["Type"])
		}
	}

	return stores
}

func (this *Factory) NewS3ImageStore(config map[string]string) imagestore.ImageStore {
	bucket := os.Getenv("S3_BUCKET")
	if len(bucket) == 0 {
		bucket = config["BucketName"]
	}

	auth, err := aws.EnvAuth()
	if err != nil {
		log.Fatal(err)
	}

	client := s3.New(auth, aws.Regions[config["Region"]])
	mapper := imagestore.NewNamePathMapper(config["NamePathRegex"], config["NamePathMap"])

	return imagestore.NewS3ImageStore(
		bucket,
		config["StoreRoot"],
		client,
		mapper,
	)
}

func (this *Factory) NewLocalImageStore(config map[string]string) imagestore.ImageStore {
	mapper := imagestore.NewNamePathMapper(config["NamePathRegex"], config["NamePathMap"])
	return imagestore.NewLocalImageStore(config["StoreRoot"], mapper)
}

func (this *Factory) NewStoreObject(name string, mime string, imgType string) *imagestore.StoreObject {
	return &imagestore.StoreObject{
		Name:     name,
		MimeType: mime,
		Type:     imgType,
	}
}

func (this *Factory) NewHashGenerator(store imagestore.ImageStore) *HashGenerator {
	hashGen := &HashGenerator{
		make(chan string),
		this.config.HashLength,
		store,
	}

	hashGen.init()
	return hashGen
}
