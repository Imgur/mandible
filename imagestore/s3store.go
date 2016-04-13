package imagestore

import (
	"io"
	"os"

	"github.com/mitchellh/goamz/s3"
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
	srcFd, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer srcFd.Close()

	bucket := this.client.Bucket(this.bucketName)

	stats, err := srcFd.Stat()
	if err != nil {
		return nil, err
	}

	err = bucket.PutReader(this.toPath(obj), srcFd, stats.Size(), obj.MimeType, s3.BucketOwnerFull)
	if err != nil {
		return nil, err
	}

	obj.Url = bucket.URL(this.toPath(obj))
	return obj, nil
}

func (this *S3ImageStore) Get(obj *StoreObject) (io.ReadCloser, error) {
	bucket := this.client.Bucket(this.bucketName)
	data, err := bucket.GetReader(this.toPath(obj))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (this *S3ImageStore) String() string {
	return "S3Store"
}

func (this *S3ImageStore) toPath(obj *StoreObject) string {
	return this.storeRoot + "/" + this.namePathMapper.mapToPath(obj)
}
