package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Starting server!")

	// Change path to location of ImgurGo config
	configFile := "config/default.config.json"
	if _, err := os.Stat(configFile); err != nil {
		fmt.Printf("Configuration file %s does not exist", configFile)
		os.Exit(-1)
	}

	config := NewConfiguration(configFile)
	server := CreateServer(config)
	server.initServer()
}
