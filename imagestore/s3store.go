package imagestore

import (
	"github.com/mitchellh/goamz/s3"
	"io/ioutil"
)

type S3ImageStore struct {
	bucketName     string
	storeRoot      string
	client         *s3.S3
	namePathMapper *NamePathMapper
}

func NewS3ImageStore(bucket string, root string, client *s3.S3, mapper *NamePathMapper) *S3ImageStore {
	return &S3ImageStore{
		bucketName:     bucket,
		storeRoot:      root,
		client:         client,
		namePathMapper: mapper,
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

func (this *S3ImageStore) Save(src string, obj *StoreObject) (*StoreObject, error) {
	bucket := this.client.Bucket(this.bucketName)

	data, err := ioutil.ReadFile(src)
	if err != nil {
		return nil, err
	}

	err = bucket.Put(this.toPath(obj), data, obj.MimeType, s3.PublicReadWrite)
	if err != nil {
		return nil, err
	}

	obj.Url = "https://s3.amazonaws.com/" + this.bucketName + this.toPath(obj)
	return obj, nil
}

func (this *S3ImageStore) toPath(obj *StoreObject) string {
	return this.storeRoot + "/" + this.namePathMapper.mapToPath(obj)
}
