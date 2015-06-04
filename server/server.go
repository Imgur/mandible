package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/Imgur/mandible/config"
	"github.com/Imgur/mandible/imageprocessor"
	"github.com/Imgur/mandible/imagestore"
	"github.com/Imgur/mandible/uploadedfile"
)

type Server struct {
	Config        *config.Configuration
	HTTPClient    *http.Client
	imageStore    imagestore.ImageStore
	hashGenerator *imagestore.HashGenerator
}

func CreateServer(c *config.Configuration) *Server {
	factory := imagestore.NewFactory(c)
	httpclient := &http.Client{}
	stores := factory.NewImageStores()
	store := stores[0]

	hashGenerator := factory.NewHashGenerator(store)
	return &Server{c, httpclient, store, hashGenerator}
}

func (s *Server) uploadFile(uploadFile io.Reader, w http.ResponseWriter, fileName string, thumbs []*uploadedfile.ThumbFile) {
	w.Header().Set("Content-Type", "application/json")

	tmpFile, err := ioutil.TempFile(os.TempDir(), "image")
	if err != nil {
		fmt.Println(err)
		ErrorResponse(w, "Unable to write to /tmp", http.StatusInternalServerError)
		return
	}

	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, uploadFile)

	if err != nil {
		fmt.Println(err)
		ErrorResponse(w, "Unable to copy image to disk!", http.StatusInternalServerError)
		return
	}

	upload, err := uploadedfile.NewUploadedFile(fileName, tmpFile.Name(), thumbs)
	defer upload.Clean()

	if err != nil {
		ErrorResponse(w, "Error detecting mime type!", http.StatusInternalServerError)
		return
	}

	processor, err := imageprocessor.Factory(s.Config.MaxFileSize, upload)
	if err != nil {
		ErrorResponse(w, "Unable to process image!", http.StatusInternalServerError)
		return
	}

	err = processor.Run(upload)
	if err != nil {
		ErrorResponse(w, "Unable to process image!", http.StatusInternalServerError)
		return
	}

	upload.SetHash(s.hashGenerator.Get())

	factory := imagestore.NewFactory(s.Config)
	obj := factory.NewStoreObject(upload.GetHash(), upload.GetMime(), "original")

	uploadFilepath := upload.GetPath()
	uploadFileFd, err := os.Open(uploadFilepath)

	if err != nil {
		ErrorResponse(w, "Unable to save image!", http.StatusInternalServerError)
	}

	obj, err = s.imageStore.Save(uploadFileFd, obj)
	if err != nil {
		ErrorResponse(w, "Unable to save image!", http.StatusInternalServerError)
		return
	}

	thumbsResp := map[string]interface{}{}
	for _, t := range upload.GetThumbs() {
		thumbName := fmt.Sprintf("%s/%s", upload.GetHash(), t.GetName())
		tObj := factory.NewStoreObject(thumbName, upload.GetMime(), "t")

		tPath := t.GetPath()
		tFile, err := os.Open(tPath)

		if err != nil {
			ErrorResponse(w, "Unable to save thumbnail!", http.StatusInternalServerError)
			return
		}

		tObj, err = s.imageStore.Save(tFile, tObj)
		if err != nil {
			ErrorResponse(w, "Unable to save thumbnail!", http.StatusInternalServerError)
			return
		}

		thumbsResp[t.GetName()] = tObj.Url
	}

	size, err := upload.FileSize()
	if err != nil {
		ErrorResponse(w, "Unable to fetch image metadata!", http.StatusInternalServerError)
		return
	}

	width, height, err := upload.Dimensions()

	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"link":   obj.Url,
		"mime":   obj.MimeType,
		"name":   fileName,
		"size":   size,
		"width":  width,
		"height": height,
		"thumbs": thumbsResp,
	}

	Response(w, resp)
}

func (s *Server) Configure(muxer *http.ServeMux) {
	fileHandler := func(w http.ResponseWriter, r *http.Request) {
		uploadFile, header, err := r.FormFile("image")

		if err != nil {
			fmt.Println(err)
			ErrorResponse(w, "Error processing file!", http.StatusInternalServerError)
			return
		}

		thumbs, err := parseThumbs(r)
		if err != nil {
			ErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.uploadFile(uploadFile, w, header.Filename, thumbs)
		uploadFile.Close()
	}

	urlHandler := func(w http.ResponseWriter, r *http.Request) {
		uploadFile, err := s.download(r.FormValue("image"))

		if err != nil {
			ErrorResponse(w, "Error dowloading URL!", http.StatusInternalServerError)
			return
		}

		thumbs, err := parseThumbs(r)
		if err != nil {
			ErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.uploadFile(uploadFile, w, "", thumbs)
		uploadFile.Close()
	}

	base64Handler := func(w http.ResponseWriter, r *http.Request) {
		b64data := r.FormValue("image")

		uploadFile := base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64data))

		thumbs, err := parseThumbs(r)
		if err != nil {
			ErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.uploadFile(uploadFile, w, "", thumbs)
	}

	rootHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<html><head><title>An open source image uploader by Imgur</title></head><body style=\"background-color: #2b2b2b; color: white\">")
		fmt.Fprint(w, "Congratulations! Your image upload server is up and running. Head over to the <a style=\"color: #85bf25 \" href=\"https://github.com/Imgur/mandible\">github</a> page for documentation")
		fmt.Fprint(w, "<br/><br/><br/><img src=\"http://i.imgur.com/YbfUjs5.png?2\" />")
		fmt.Fprint(w, "</body></html>")
	}

	muxer.HandleFunc("/file", fileHandler)
	muxer.HandleFunc("/url", urlHandler)
	muxer.HandleFunc("/base64", base64Handler)
	muxer.HandleFunc("/", rootHandler)
}

func (s *Server) download(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", s.Config.UserAgent)

	resp, err := s.HTTPClient.Do(req)

	if err != nil {
		// "HTTP protocol error" - maybe the server sent an invalid response or timed out
		return nil, err
	}

	if 200 != resp.StatusCode {
		return nil, errors.New("Non-200 status code received")
	}

	contentLength := resp.ContentLength

	if contentLength == 0 {
		return nil, errors.New("Empty file received")
	}

	return resp.Body, nil
}

func parseThumbs(r *http.Request) ([]*uploadedfile.ThumbFile, error) {
	thumbString := r.FormValue("thumbs")
	if thumbString == "" {
		return []*uploadedfile.ThumbFile{}, nil
	}

	var t map[string]map[string]interface{}
	err := json.Unmarshal([]byte(thumbString), &t)
	if err != nil {
		return nil, errors.New("Error parsing thumbnail JSON!")
	}

	var thumbs []*uploadedfile.ThumbFile
	for name, thumb := range t {
		width, wOk := thumb["width"].(float64)
		if !wOk {
			return nil, errors.New("Invalid thumbnail width!")
		}

		height, hOk := thumb["height"].(float64)
		if !hOk {
			return nil, errors.New("Invalid thumbnail height!")
		}

		shape, sOk := thumb["shape"].(string)
		if !sOk {
			return nil, errors.New("Invalid thumbnail shape!")
		}

		switch shape {
		case "thumb":
		case "square":
		case "circle":
		default:
			return nil, errors.New("Invalid thumbnail shape!")
		}

		thumbs = append(thumbs, uploadedfile.NewThumbFile(int(width), int(height), name, shape, ""))
	}

	return thumbs, nil
}
