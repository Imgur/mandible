package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

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
	b64bytes, _ := base64.StdEncoding.DecodeString(b64gif)

	values := make(url.Values)
	values.Add("image", b64gif)

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

	if !*serverResp.Success {
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

	immStore := server.ImageStore
	exists, err := immStore.Exists(&imagestore.StoreObject{Id: imageResp.Hash})
	if err != nil {
		t.Fatalf("Unexpected error checking if %s exists in in-memory image store: %s", imageResp.Hash, err.Error())
	}

	if !exists {
		t.Fatalf("Expected to find %s in the in-memory storage, instead absent. Dump: %+v", imageResp.Hash, immStore)
	}

	storedBodyReader, err := immStore.Get(&imagestore.StoreObject{Id: imageResp.Hash})
	if err != nil {
		t.Fatalf("Unexpected error fetching %s from in-memory image store: %s", imageResp.Hash, err.Error())
	}
	storedBodyBytes, _ := ioutil.ReadAll(storedBodyReader)

	if !bytes.Equal(storedBodyBytes, []byte(b64bytes)) {
		t.Fatalf("Stored bytes %s != %s", storedBodyBytes, []byte(b64bytes))
	}
}

func TestAuthentication(t *testing.T) {
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
	authenticator := NewHMACAuthenticatorSHA256([]byte("foobar"))
	server := NewAuthenticatedServer(cfg, imageprocessor.PassthroughStrategy, authenticator)

	muxer := http.NewServeMux()

	server.Configure(muxer)

	ts := httptest.NewServer(muxer)
	defer ts.Close()

	values := make(url.Values)
	values.Add("image", b64gif)

	req, err := http.NewRequest("POST", ts.URL+"/user/123/base64", strings.NewReader(values.Encode()))
	if err != nil {
		t.Fatalf("Error when forming authenticated base64 GIF upload request: %s", err.Error())
	}

	message := AuthenticatedUser{
		UserID:               "123",
		GrantTime:            time.Now(),
		GrantDurationSeconds: 365 * 24 * 3600,
	}
	messageBytes, _ := json.Marshal(&message)
	messageMacWriter := hmac.New(sha256.New, []byte("foobar"))
	messageMacWriter.Write(messageBytes)
	messageMac := base64.StdEncoding.EncodeToString(messageMacWriter.Sum(nil))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", string(messageBytes))
	req.Header.Set("X-Authorization-HMAC", string(messageMac))

	httpclient := http.Client{}
	res, err := httpclient.Do(req)
	if err != nil {
		t.Fatalf("Error when uploading authenticated base64 GIF: %s", err.Error())
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

	if !*serverResp.Success {
		t.Fatalf("Uploading GIF was unsuccessful")
	}

	imageRespBytes, _ := json.Marshal(serverResp.Data)

	err = json.Unmarshal(imageRespBytes, &imageResp)

	if imageResp.Mime != "image/gif" {
		t.Fatalf("Expected image MIME type to be image/gif, instead %s", imageResp.Mime)
	}

	if imageResp.UserID != "123" {
		t.Fatalf("Expected user ID to be \"123\", instead \"%s\"", imageResp.UserID)
	}
}

func TestGetFullWebpThumb(t *testing.T) {
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
	server := NewServer(cfg, imageprocessor.ThumbnailStrategy)
	muxer := http.NewServeMux()
	server.Configure(muxer)
	ts := httptest.NewServer(muxer)
	defer ts.Close()

	thumbsJson, _ := json.Marshal(map[string]interface{}{
		"webp": map[string]interface{}{
			"format": "webp",
		},
	})

	values := make(url.Values)
	values.Add("image", b64dan)
	values.Add("thumbs", string(thumbsJson))

	res, err := http.PostForm(ts.URL+"/base64", values)
	if err != nil {
		t.Fatalf("Error when uploading base64 image: %s", err.Error())
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

	if !*serverResp.Success {
		t.Fatalf("Uploading image was unsuccessful")
	}

	imageRespBytes, _ := json.Marshal(serverResp.Data)

	err = json.Unmarshal(imageRespBytes, &imageResp)

	if len(imageResp.Thumbs) == 0 {
		t.Fatalf("Expected thumbs to contain data, instead blank")
	}

	if _, ok := imageResp.Thumbs["webp"]; !ok {
		t.Fatalf("Expected webp thumb, not given")
	}

	immStore := server.ImageStore
	storeId := imageResp.Hash + "/webp"

	exists, err := immStore.Exists(&imagestore.StoreObject{Id: storeId})
	if err != nil {
		t.Fatalf("Unexpected error checking if %s exists in in-memory image store: %s", storeId, err.Error())
	}
	if !exists {
		t.Fatalf("Expected to find %s in the in-memory storage, instead absent. Dump: %+v", storeId, immStore)
	}

	storedBodyReader, err := immStore.Get(&imagestore.StoreObject{Id: storeId})
	if err != nil {
		t.Fatalf("Unexpected error fetching %s from in-memory image store: %s", storeId, err.Error())
	}
	storedBodyBytes, _ := ioutil.ReadAll(storedBodyReader)

	if len(storedBodyBytes) == 0 {
		t.Fatalf("Expected webp thumbnail to be larger than 0 bytes")
	}

	if int64(len(storedBodyBytes)) >= imageResp.Size {
		t.Fatalf("Expected thumbnail to be smaller than original image, %v vs %v", int64(len(storedBodyBytes)), imageResp.Size)
	}
}

func TestGetSizedWebpThumb(t *testing.T) {
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
	server := NewServer(cfg, imageprocessor.ThumbnailStrategy)
	muxer := http.NewServeMux()
	server.Configure(muxer)
	ts := httptest.NewServer(muxer)
	defer ts.Close()

	thumbsJson, _ := json.Marshal(map[string]interface{}{
		"webp": map[string]interface{}{
			"format": "webp",
		},
		"webpthumb": map[string]interface{}{
			"format": "webp",
			"shape":  "custom",
			"width":  "10",
			"height": "10",
		},
	})

	values := make(url.Values)
	values.Add("image", b64dan)
	values.Add("thumbs", string(thumbsJson))

	res, err := http.PostForm(ts.URL+"/base64", values)
	if err != nil {
		t.Fatalf("Error when uploading base64 iamge: %s", err.Error())
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

	if !*serverResp.Success {
		t.Fatalf("Uploading image was unsuccessful")
	}

	imageRespBytes, _ := json.Marshal(serverResp.Data)

	err = json.Unmarshal(imageRespBytes, &imageResp)

	if len(imageResp.Thumbs) == 0 {
		t.Fatalf("Expected thumbs to contain data, instead blank")
	}

	if _, ok := imageResp.Thumbs["webp"]; !ok {
		t.Fatalf("Expected webp thumb, not given")
	}

	immStore := server.ImageStore
	storeId := imageResp.Hash + "/webp"
	storeIdSmall := imageResp.Hash + "/webpthumb"

	exists, err := immStore.Exists(&imagestore.StoreObject{Id: storeIdSmall})
	if err != nil {
		t.Fatalf("Unexpected error checking if %s exists in in-memory image store: %s", storeIdSmall, err.Error())
	}
	if !exists {
		t.Fatalf("Expected to find %s in the in-memory storage, instead absent. Dump: %+v", storeIdSmall, immStore)
	}

	storedBodyReader, err := immStore.Get(&imagestore.StoreObject{Id: storeId})
	if err != nil {
		t.Fatalf("Unexpected error fetching %s from in-memory image store: %s", storeId, err.Error())
	}
	storedBodyReaderSmall, err := immStore.Get(&imagestore.StoreObject{Id: storeIdSmall})
	if err != nil {
		t.Fatalf("Unexpected error fetching %s from in-memory image store: %s", storeIdSmall, err.Error())
	}
	storedBodyBytes, _ := ioutil.ReadAll(storedBodyReader)
	storedBodyBytesSmall, _ := ioutil.ReadAll(storedBodyReaderSmall)

	if len(storedBodyBytesSmall) == 0 {
		t.Fatalf("Expected webp thumbnail to be larger than 0 bytes")
	}

	if len(storedBodyBytesSmall) >= len(storedBodyBytes) {
		t.Fatalf("Expected thumbnail to be smaller than original image, %v vs %v", len(storedBodyBytesSmall), len(storedBodyBytes))
	}
}

func TestTooLarge(t *testing.T) {
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
	server := NewServer(cfg, imageprocessor.ThumbnailStrategy)
	muxer := http.NewServeMux()
	server.Configure(muxer)
	ts := httptest.NewServer(muxer)
	defer ts.Close()

	thumbsJson, _ := json.Marshal(map[string]interface{}{
		"webp": map[string]interface{}{
			"format": "webp",
			"shape":  "custom",
			"width":  "20000",
			"height": "20000",
		},
	})

	values := make(url.Values)
	values.Add("image", b64dan)
	values.Add("thumbs", string(thumbsJson))

	res, err := http.PostForm(ts.URL+"/base64", values)
	if err != nil {
		t.Fatalf("Error when uploading base64 iamge: %s", err.Error())
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
	err = json.Unmarshal(body, &serverResp)
	if err != nil {
		t.Fatalf("Unexpected error parsing response: %s", err.Error())
	}

	t.Logf("%v+", serverResp)
	if *serverResp.Success {
		t.Fatalf("Uploading large image was successful")
	}
}

func TestTooSmall(t *testing.T) {
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
	server := NewServer(cfg, imageprocessor.ThumbnailStrategy)
	muxer := http.NewServeMux()
	server.Configure(muxer)
	ts := httptest.NewServer(muxer)
	defer ts.Close()

	thumbsJson, _ := json.Marshal(map[string]interface{}{
		"webp": map[string]interface{}{
			"format": "webp",
			"shape":  "custom",
			"width":  0,
			"height": 0,
		},
	})

	values := make(url.Values)
	values.Add("image", b64dan)
	values.Add("thumbs", string(thumbsJson))

	res, err := http.PostForm(ts.URL+"/base64", values)
	if err != nil {
		t.Fatalf("Error when uploading base64 image: %s", err.Error())
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
	err = json.Unmarshal(body, &serverResp)
	if err != nil {
		t.Fatalf("Unexpected error parsing response: %s", err.Error())
	}

	if *serverResp.Success {
		t.Fatalf("Uploading small image was successful")
	}
}

var (
	b64gif = "R0lGODlhAQABAIAAAAAAAP" + "/" + "/" + "/yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"
	b64dan = "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"
)
