package main

import (
	"fmt"
	//"os"
)

func main() {
	// Change path to location of ImgurGo config
	configFile := "config/default.config.json"
    //if _, err := os.Stat(configFile); err != nil {
	//	fmt.Printf("Configuration file %s does not exist", configFile)
	//	os.Exit(-1)
	//}
    fmt.Printf("got here")
	config := NewConfiguration(configFile)
    fmt.Printf("got here2")
	server := CreateServer(config)
    fmt.Printf("got here3")
	server.initServer()
    fmt.Printf("got here4")
}
