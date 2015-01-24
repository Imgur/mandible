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
	file, err := os.Open(path)

	if err != nil {
		fmt.Printf("Error leading config file!")
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
