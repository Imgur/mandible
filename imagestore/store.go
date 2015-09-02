package imagestore

import (
	"fmt"
	"io"
)

type ImageStore interface {
	Save(src string, obj *StoreObject) (*StoreObject, error)
	Exists(obj *StoreObject) (bool, error)
	Get(obj *StoreObject) (io.ReadCloser, error)
	String() string
}

type MultiImageStore []ImageStore

func (this MultiImageStore) Save(src string, obj *StoreObject) (*StoreObject, error) {
	errs := make(chan error, len(this))

	for _, store := range this {
		go func(s ImageStore) {
			_, err := s.Save(src, obj)
			if err != nil {
				errs <- fmt.Errorf("Error asynchronously saving image on %s: %s", s.String(), err.Error())
			} else {
				errs <- nil
			}
		}(store)
	}

	for i := 0; i < len(this); i++ {
		select {
		case err := <-errs:
			if err != nil {
				return nil, err
			}
		}
	}

	return obj, nil
}

func (this MultiImageStore) Exists(obj *StoreObject) (bool, error) {
	errs := make(chan error, len(this))
	results := make(chan bool, len(this))

	for _, store := range this {
		go func(s ImageStore) {
			r, err := s.Exists(obj)
			if err != nil {
				errs <- fmt.Errorf("Error asynchronously proving existance for image on %s: %s", s.String(), err.Error())
			} else {
				results <- r
			}
		}(store)
	}

	for i := 0; i < len(this); i++ {
		select {
		case err := <-errs:
			if err != nil {
				return false, err
			}
		case r := <-results:
			if r == true {
				return true, nil
			}
		}
	}

	return false, nil
}

func (this MultiImageStore) Get(obj *StoreObject) (io.ReadCloser, error) {
	errs := make(chan error, len(this))
	results := make(chan io.ReadCloser, 1)
	done := make(chan bool, 1)

	for _, store := range this {
		go func(s ImageStore) {
			r, err := s.Get(obj)
			if err != nil {
				errs <- fmt.Errorf("Error asynchronously getting image on %s: %s", s.String(), err.Error())
			} else {
				select {
				case done <- true:
					results <- r
				default:
					r.Close()
				}
			}
		}(store)
	}

	var err error

	for i := 0; i < len(this); i++ {
		select {
		case r := <-results:
			return r, nil
		case err = <-errs:
		}
	}

	return nil, err
}

func (this MultiImageStore) String() string {
	str := ""

	for _, store := range this {
		str += store.String()
		str += " "
	}

	return str
}
