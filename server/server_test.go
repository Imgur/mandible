package server

import (
	"github.com/Imgur/mandible/config"
	"github.com/Imgur/mandible/imagestore"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestingTheFrontPageGetsSomeHTML(t *testing.T) {
	memorystore := imagestore.NewInMemoryImageStore()

	cfg := &config.Configuration{
		MaxFileSize: 99999999999,
		HashLength:  7,
		UserAgent:   "Foobar",
		Stores:      make([]map[string]string, 0),
		Port:        8888,
	}

	factory := imagestore.NewFactory(cfg)
	hasher := factory.NewHashGenerator(memorystore)

	server := &Server{
		Config:        cfg,
		HTTPClient:    &http.Client{},
		imageStore:    memorystore,
		hashGenerator: hasher,
	}

	muxer := http.NewServeMux()

	server.Configure(muxer)

	ts := httptest.NewServer(muxer)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Error when retrieving %s: %s", ts.URL, err.Error())
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Failed to read response body of %s: %s", ts.URL, err.Error())
	}

	t.Logf("Response to %s/ was: %s", ts.URL, body)

	if res.StatusCode != 200 {
		t.Fatalf("Unexpected status code %d", res.StatusCode)
	}

	sbody := string(body)

	if !strings.Contains(sbody, "<html>") {
		t.Fatalf("Did I get HTML back? Didn't find <html>...")
	}
}
