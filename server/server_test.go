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
	stats := &DiscardStats{}
	server := NewServer(cfg, imageprocessor.PassthroughStrategy, stats)

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
	stats := &DiscardStats{}
	server := NewServer(cfg, imageprocessor.PassthroughStrategy, stats)

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
	stats := &DiscardStats{}
	server := NewAuthenticatedServer(cfg, imageprocessor.PassthroughStrategy, authenticator, stats)

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
	stats := &DiscardStats{}
	server := NewServer(cfg, imageprocessor.ThumbnailStrategy, stats)
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
	stats := &DiscardStats{}
	server := NewServer(cfg, imageprocessor.ThumbnailStrategy, stats)
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
			"width":  10,
			"height": 10,
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
	stats := &DiscardStats{}
	server := NewServer(cfg, imageprocessor.ThumbnailStrategy, stats)
	muxer := http.NewServeMux()
	server.Configure(muxer)
	ts := httptest.NewServer(muxer)
	defer ts.Close()

	thumbsJson, _ := json.Marshal(map[string]interface{}{
		"webp": map[string]interface{}{
			"format": "webp",
			"shape":  "custom",
			"width":  20000,
			"height": 20000,
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

	if res.StatusCode != 500 {
		t.Fatalf("Unexpected status code %d", res.StatusCode)
	}

	var serverResp ServerResponse
	err = json.Unmarshal(body, &serverResp)
	if err != nil {
		t.Fatalf("Unexpected error parsing response: %s", err.Error())
	}

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
	stats := &DiscardStats{}
	server := NewServer(cfg, imageprocessor.ThumbnailStrategy, stats)
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

	if res.StatusCode != 500 {
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

func TestGetTallThumb(t *testing.T) {
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
	stats := &DiscardStats{}
	server := NewServer(cfg, imageprocessor.ThumbnailStrategy, stats)
	muxer := http.NewServeMux()
	server.Configure(muxer)
	ts := httptest.NewServer(muxer)
	defer ts.Close()

	thumbsJson, _ := json.Marshal(map[string]interface{}{
		"tallthumb": map[string]interface{}{
			"shape":        "custom",
			"crop_gravity": "north",
			"crop_ratio":   "1:2.25",
			"max_width":    10,
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

	if _, ok := imageResp.Thumbs["tallthumb"]; !ok {
		t.Fatalf("Expected cropped thumb, not given")
	}

	immStore := server.ImageStore
	storeId := imageResp.Hash
	storeIdSmall := imageResp.Hash + "/tallthumb"

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

var (
	b64gif = "R0lGODlhAQABAIAAAAAAAP" + "/" + "/" + "/yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"
	b64dan = "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"
)
