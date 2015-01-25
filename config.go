package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
	MaxFileSize int64
	HashLength  int
	UserAgent   string
	Stores      []StoreConfigWrapper
	Port        int
}

type BaseStoreConfig struct {
	StoreRoot     string
	NamePathRegex string
	NamePathMap   string
}

type S3StoreConfig struct {
	*BaseStoreConfig
	BucketName string
	Region     string
}

type LocalStoreConfig struct {
	*BaseStoreConfig
}

type StoreConfigWrapper struct {
	Type   string
	Config interface{}
}

func NewConfiguration(path string) *Configuration {
	file, err := os.Open(path)

	if err != nil {
		fmt.Printf("Error opening config file!")
		os.Exit(-1)
	}

	decoder := json.NewDecoder(file)
	configuration := &Configuration{}
	err = decoder.Decode(configuration)

	if err != nil {
		fmt.Println("Error loading config file: ", err)
	}

	return configuration
}
