package imagestore

import (
	"fmt"
	"github.com/ernado/selectel/storage"
	"io"
	"os"
	"path"
)

type SelectelStore struct {
	client         storage.API
	storeRoot      string
	container      string
	namePathMapper *NamePathMapper
}

func NewSelectelImageStore(client storage.API, mapper *NamePathMapper, container, root string) *SelectelStore {
	return &SelectelStore{
		client:         client,
		namePathMapper: mapper,
		container:      container,
		storeRoot:      root,
	}
}

func (s *SelectelStore) Save(src string, obj *StoreObject) (*StoreObject, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pathToFile := s.toPath(obj)
	container, name := s.toSelectelPath(pathToFile)
	if err := s.client.Upload(f, container, name, obj.MimeType); err != nil {
		return nil, fmt.Errorf("Selectel api returns error: %v", err)
	}

	obj.Url = s.client.URL(container, name)

	return obj, nil
}
func (s *SelectelStore) Exists(obj *StoreObject) (bool, error) {
	pathToFile := s.toPath(obj)
	container, name := s.toSelectelPath(pathToFile)
	_, err := s.client.ObjectInfo(container, name)
	return err == nil, nil
}
func (s *SelectelStore) Get(obj *StoreObject) (io.ReadCloser, error) {
	pathToFile := s.toPath(obj)
	container, name := s.toSelectelPath(pathToFile)
	return s.client.C(container).Object(name).GetReader()
}
func (s *SelectelStore) String() string {
	return "SelectelStore"
}

func (s *SelectelStore) toPath(obj *StoreObject) string {
	return s.storeRoot + "/" + s.namePathMapper.mapToPath(obj)
}

func (s *SelectelStore) toSelectelPath(fullPath string) (string, string) {
	return path.Join(s.container, path.Dir(fullPath)), path.Base(fullPath)
}
