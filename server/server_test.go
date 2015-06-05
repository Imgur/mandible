package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Imgur/mandible/config"
	"github.com/Imgur/mandible/imageprocessor"
	"github.com/Imgur/mandible/imagestore"
)

func TestRequestingTheFrontPageGetsSomeHTML(t *testing.T) {
	cfg := &config.Configuration{
		MaxFileSize: 99999999999,
		HashLength:  7,
		UserAgent:   "Foobar",
		Stores:      make([]map[string]string, 0),
		Port:        8888,
	}

	memcfg := make(map[string]string)
	memcfg["Type"] = "memory"
	cfg.Stores = append(cfg.Stores, memcfg)
	server := NewServer(cfg, imageprocessor.PassthroughStrategy)

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

func TestPostingBase64FilePutsTheFileInStorageAndReturnsJSON(t *testing.T) {
	cfg := &config.Configuration{
		MaxFileSize: 99999999999,
		HashLength:  7,
		UserAgent:   "Foobar",
		Stores:      make([]map[string]string, 0),
		Port:        8888,
	}

	memcfg := make(map[string]string)
	memcfg["Type"] = "memory"
	cfg.Stores = append(cfg.Stores, memcfg)
	server := NewServer(cfg, imageprocessor.PassthroughStrategy)

	muxer := http.NewServeMux()

	server.Configure(muxer)

	ts := httptest.NewServer(muxer)
	defer ts.Close()

	// a 1x1 base64 encoded transparent GIF
	b64gif := "R0lGODlhAQABAIAAAAAAAP" + "/" + "/" + "/yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"
	b64bytes, _ := base64.StdEncoding.DecodeString(b64gif)

	values := make(url.Values)
	values["image"] = make([]string, 0)
	values["image"] = append(values["image"], b64gif)

	res, err := http.PostForm(ts.URL+"/base64", values)
	if err != nil {
		t.Fatalf("Error when uploading base64 GIF: %s", err.Error())
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %s", err.Error())
	}

	t.Logf("Response to /base64 was: %s", body)

	if res.StatusCode != 200 {
		t.Fatalf("Unexpected status code %d", res.StatusCode)
	}

	var serverResp ServerResponse
	var imageResp ImageResponse
	err = json.Unmarshal(body, &serverResp)
	if err != nil {
		t.Fatalf("Unexpected error parsing response: %s", err.Error())
	}

	if !serverResp.Success {
		t.Fatalf("Uploading GIF was unsuccessful")
	}

	imageRespBytes, _ := json.Marshal(serverResp.Data)

	err = json.Unmarshal(imageRespBytes, &imageResp)

	if imageResp.Height != 1 {
		t.Fatalf("Expected height to be 1, instead %d", imageResp.Height)
	}

	if imageResp.Width != 1 {
		t.Fatalf("Expected width to be 1, instead %d", imageResp.Width)
	}

	if imageResp.Size != 42 {
		t.Fatalf("Expected size to be 42, instead %d", imageResp.Size)
	}

	if imageResp.Mime != "image/gif" {
		t.Fatalf("Expected image MIME type to be image/gif, instead %s", imageResp.Mime)
	}

	immStore := server.ImageStore.(*imagestore.InMemoryImageStore)
	exists, err := immStore.Exists(&imagestore.StoreObject{Name: imageResp.Hash})
	if err != nil {
		t.Fatalf("Unexpected error checking if %s exists in in-memory image store: %s", imageResp.Hash, err.Error())
	}

	if !exists {
		t.Fatalf("Expected to find %s in the in-memory storage, instead absent. Dump: %+v", imageResp.Hash, immStore)
	}

	storedBodyReader, err := immStore.Get(&imagestore.StoreObject{Name: imageResp.Hash})
	if err != nil {
		t.Fatalf("Unexpected error fetching %s from in-memory image store: %s", imageResp.Hash, err.Error())
	}
	storedBodyBytes, _ := ioutil.ReadAll(storedBodyReader)

	if !bytes.Equal(storedBodyBytes, []byte(b64bytes)) {
		t.Fatalf("Stored bytes %s != %s", storedBodyBytes, []byte(b64bytes))
	}
}
