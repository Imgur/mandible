package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
	MaxFileSize     int64
	HashLength      int
	UserAgent       string
	Stores          []map[string]string
	Port            int
	DatadogEnabled  bool
	DatadogHostname string
	LabelingEnabled bool
	LabelModelDir   string
	LabelTopN       int
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
