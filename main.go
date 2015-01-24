package main

import (
	"fmt"
	"os"
)

func main() {
    configFile := os.Getenv("IMGUR_GO_CONF")

    if _, err := os.Stat(configFile); err != nil {
		fmt.Printf("Configuration file %s does not exist", configFile)
		os.Exit(-1)
	}
	
    config := NewConfiguration(configFile)
	server := CreateServer(config)
	server.initServer()
}
