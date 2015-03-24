package imagestore

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/Imgur/mandible/config"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gcloud "google.golang.org/cloud"
	gcs "google.golang.org/cloud/storage"
)

type Factory struct {
	conf *config.Configuration
}

func NewFactory(conf *config.Configuration) *Factory {
	return &Factory{conf}
}

func (this *Factory) NewImageStores() []ImageStore {
	stores := []ImageStore{}

	for _, configWrapper := range this.conf.Stores {
		switch configWrapper["Type"] {
		case "s3":
			store := this.NewS3ImageStore(configWrapper)
			stores = append(stores, store)
		case "gcs":
			store := this.NewGCSImageStore(configWrapper)
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

func (this *Factory) NewS3ImageStore(conf map[string]string) ImageStore {
	bucket := os.Getenv("S3_BUCKET")
	if len(bucket) == 0 {
		bucket = conf["BucketName"]
	}

	auth, err := aws.EnvAuth()
	if err != nil {
		log.Fatal(err)
	}

	client := s3.New(auth, aws.Regions[conf["Region"]])
	mapper := NewNamePathMapper(conf["NamePathRegex"], conf["NamePathMap"])

	return NewS3ImageStore(
		bucket,
		conf["StoreRoot"],
		client,
		mapper,
	)
}

func (this *Factory) NewGCSImageStore(conf map[string]string) ImageStore {
	jsonKey, err := ioutil.ReadFile(conf["KeyFile"])
	if err != nil {
		log.Fatal(err)
	}
	cloudConf, err := google.JWTConfigFromJSON(
		jsonKey,
		gcs.ScopeFullControl,
	)
	if err != nil {
		log.Fatal(err)
	}

	bucket := os.Getenv("GCS_BUCKET")
	if len(bucket) == 0 {
		bucket = conf["BucketName"]
	}

	ctx := gcloud.NewContext(conf["AppID"], cloudConf.Client(oauth2.NoContext))
	mapper := NewNamePathMapper(conf["NamePathRegex"], conf["NamePathMap"])

	return NewGCSImageStore(
		ctx,
		bucket,
		conf["StoreRoot"],
		mapper,
	)
}

func (this *Factory) NewLocalImageStore(conf map[string]string) ImageStore {
	mapper := NewNamePathMapper(conf["NamePathRegex"], conf["NamePathMap"])
	return NewLocalImageStore(conf["StoreRoot"], mapper)
}

func (this *Factory) NewStoreObject(name string, mime string, imgType string) *StoreObject {
	return &StoreObject{
		Name:     name,
		MimeType: mime,
		Type:     imgType,
	}
}

func (this *Factory) NewHashGenerator(store ImageStore) *HashGenerator {
	hashGen := &HashGenerator{
		make(chan string),
		this.conf.HashLength,
		store,
	}

	hashGen.init()
	return hashGen
}
