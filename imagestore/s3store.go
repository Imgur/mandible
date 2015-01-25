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
	region     string
	client     *s3.S3
}

func NewS3ImageStore(bucket string, root string, region string) *S3ImageStore {
	auth, err := aws.EnvAuth()
	if err != nil {
		log.Fatal(err)
	}

	client := s3.New(auth, aws.Regions[region])
	return &S3ImageStore{
		bucketName: bucket,
		storeRoot:  root,
		region:     region,
		client:     client,
	}
}

func (this *S3ImageStore) Exists(obj *StoreObject) (bool, error) {
	bucket := this.client.Bucket(this.bucketName)
	response, err := bucket.Head(this.toPath(obj))

	if err != nil {
		return false, err
	}

	return (response.StatusCode == 200), nil
}

func (this *S3ImageStore) Save(src string, obj *StoreObject) error {
	bucket := this.client.Bucket(this.bucketName)

	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = bucket.Put(this.toPath(obj), data, obj.MimeType, s3.PublicReadWrite)
	if err != nil {
		return err
	}

	return nil
}

func (this *S3ImageStore) toPath(obj *StoreObject) string {
	return this.storeRoot + "/" + obj.Type + "/" + obj.Name
}
