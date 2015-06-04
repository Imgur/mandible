package main

import (
	"os"

	mandibleConf "github.com/Imgur/mandible/config"
	mandible "github.com/Imgur/mandible/server"
)

func main() {
	configFile := os.Getenv("IMGUR_GO_CONF")

	config := mandibleConf.NewConfiguration(configFile)
	server := mandible.CreateServer(config)
	server.Start()
}
