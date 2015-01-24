package main

import (
	"fmt"
	//"os"
)

func main() {
	// Change path to location of ImgurGo config
	configFile := "config/default.conf.json"
    //if _, err := os.Stat(configFile); err != nil {
	//	fmt.Printf("Configuration file %s does not exist", configFile)
	//	os.Exit(-1)
	//}
    fmt.Println("got here")
	config := NewConfiguration(configFile)
    fmt.Println("got here2")
	server := CreateServer(config)
    fmt.Println("got here3")
	server.initServer()
    fmt.Println("got here4")
}
