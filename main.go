package main

import (
	"os"
    mandible "github.com/Imgur/mandible/server"
    mandibleConf "github.com/Imgur/mandible/config"
)

func main() {
	configFile := os.Getenv("IMGUR_GO_CONF")

	config := mandibleConf.NewConfiguration(configFile)
	server := mandible.CreateServer(config)
	server.Start()
}
