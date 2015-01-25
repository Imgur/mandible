package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
	MaxFileSize int64
	UserAgent   string
	Store       StoreConfig
}

type StoreConfig struct {
	S3    S3StoreConfig
	Local LocalStoreConfig
}

type S3StoreConfig struct {
	BucketName string
	Region     string
	StoreRoot  string
}

type LocalStoreConfig struct {
	StoreRoot string
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
