package imagestore

import (
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"io/ioutil"
	"log"
)

type S3ImageStore struct {
	bucketName string
	storeRoot  string
	client     s3.S3
}

func NewS3ImageStore(bucket string, root string) *S3ImageStore {
	return &S3ImageStore{
		bucketName: bucket,
		storeRoot:  root,
	}
}

func (this *S3ImageStore) Exists(obj *StoreObject) bool {
	return true
}

func (this *S3ImageStore) Save(src string, obj *StoreObject) error {
	auth, err := aws.EnvAuth()
	if err != nil {
		log.Fatal(err)
	}

	client := s3.New(auth, aws.USEast)
	bucket := client.Bucket(this.bucketName)

	data, err := ioutil.ReadFile(src)
	if err != nil {
		log.Fatal(err)
	}

	err = bucket.Put(this.toPath(obj.Path), data, obj.MimeType, s3.PublicReadWrite)

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (this *S3ImageStore) toPath(path string) string {
	return this.storeRoot + "/" + path
}
