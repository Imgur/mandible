package main

import (
	"fmt"
	"net/http"
	"os"

	mandibleConf "github.com/Imgur/mandible/config"
	mandible "github.com/Imgur/mandible/server"
)

func main() {
	configFile := os.Getenv("IMGUR_GO_CONF")

	config := mandibleConf.NewConfiguration(configFile)
	server := mandible.CreateServer(config)
	muxer := http.NewServeMux()
	server.Configure(muxer)

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = fmt.Sprintf(":%d", server.Config.Port)
	}

	http.ListenAndServe(port, muxer)
}
