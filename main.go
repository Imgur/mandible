package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	mandibleConf "github.com/Imgur/mandible/config"
	processors "github.com/Imgur/mandible/imageprocessor"
	mandible "github.com/Imgur/mandible/server"
)

func main() {
	configFile := os.Getenv("MANDIBLE_CONF")

	config := mandibleConf.NewConfiguration(configFile)

	var server *mandible.Server
	var stats mandible.RuntimeStats

	if config.DatadogEnabled {
		var err error
		stats, err = mandible.NewDatadogStats(config.DatadogHostname)
		if err != nil {
			log.Printf("Invalid Datadog Hostname: %s", config.DatadogHostname)
			os.Exit(1)
		}
		log.Println("Stats init success")
	} else {
		stats = &mandible.DiscardStats{}
	}

	if os.Getenv("AUTHENTICATION_HMAC_KEY") != "" {
		key := []byte(os.Getenv("AUTHENTICATION_HMAC_KEY"))
		auth := mandible.NewHMACAuthenticatorSHA256(key)
		server = mandible.NewAuthenticatedServer(config, processors.EverythingStrategy, auth, stats)
	} else {
		server = mandible.NewServer(config, processors.EverythingStrategy, stats)
	}

	muxer := http.NewServeMux()
	server.Configure(muxer)

	port := fmt.Sprintf(":%d", server.Config.Port)

	log.Printf("Listening on Port: %s", port)

	http.ListenAndServe(port, muxer)
}
