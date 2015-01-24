package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
	MaxFileSize int
}

func NewConfiguration(path string) *Configuration {
	file, _ := os.Open(path)
	decoder := json.NewDecoder(file)
	configuration := &Configuration{}
	err := decoder.Decode(configuration)

	if err != nil {
		fmt.Println("Error loading config file: ", err)
	}

	return configuration
}
