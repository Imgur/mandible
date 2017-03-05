package imagestore

import (
	"io/ioutil"
	"log"

	"github.com/Imgur/mandible/config"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"

	selectel "github.com/ernado/selectel/storage"
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

func (this *Factory) NewImageStores() ImageStore {
	stores := MultiImageStore{}
	var store ImageStore

	for _, configWrapper := range this.conf.Stores {
		switch configWrapper["Type"] {
		case "s3":
			store = this.NewS3ImageStore(configWrapper)
			stores = append(stores, store)
		case "gcs":
			store = this.NewGCSImageStore(configWrapper)
			stores = append(stores, store)
		case "local":
			store = this.NewLocalImageStore(configWrapper)
			stores = append(stores, store)
		case "memory":
			store = NewInMemoryImageStore()
			stores = append(stores, store)
		case "selectel":
			store = this.NewSelectelStore(configWrapper)
			stores = append(stores, store)
		default:
			log.Fatalf("Unsupported store %s", configWrapper["Type"])
		}
	}

	if len(this.conf.Stores) == 1 {
		return store
	}

	// return a MultiImageStore type if more then 1 store was specified in the config
	return stores
}

func (this *Factory) NewS3ImageStore(conf map[string]string) ImageStore {
	bucket := conf["BucketName"]

	auth, err := aws.GetAuth(conf["AWSKey"], conf["AWSSecret"])
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

	bucket := conf["BucketName"]

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

func (this *Factory) NewStoreObject(id string, mime string, size string) *StoreObject {
	return &StoreObject{
		Id:       id,
		MimeType: mime,
		Size:     size,
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

func (this *Factory) NewSelectelStore(conf map[string]string) ImageStore {
	user, key, container, rootPath := conf["user"], conf["key"], conf["container"], conf["rootPath"]
	client, err := selectel.New(user, key)
	if err != nil {
		log.Fatal(err)
	}

	mapper := NewNamePathMapper(conf["NamePathRegex"], conf["NamePathMap"])

	return NewSelectelImageStore(client, mapper, container, rootPath)
}
