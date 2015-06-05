package main

import (
	"fmt"
	"net/http"
	"os"

	mandibleConf "github.com/Imgur/mandible/config"
	processors "github.com/Imgur/mandible/imageprocessor"
	mandible "github.com/Imgur/mandible/server"
)

func main() {
	configFile := os.Getenv("IMGUR_GO_CONF")

	config := mandibleConf.NewConfiguration(configFile)
	server := mandible.NewServer(config, processors.EverythingStrategy)
	muxer := http.NewServeMux()
	server.Configure(muxer)

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = fmt.Sprintf(":%d", server.Config.Port)
	}

	http.ListenAndServe(port, muxer)
}
