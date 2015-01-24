package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

type URLUpload struct {
	url string
}

type Server struct {
	Config *Configuration
}

func CreateServer(c *Configuration) *Server {
	return &Server{c}
}

func (s *Server) initServer() {
	uploadHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		uploadType := r.FormValue("type")

		var uploadFile multipart.File
		var err error

		if uploadType == "url" {
			uploadFile, _, err = s.download(r.FormValue("image"))
		} else if uploadType == "file" {
			uploadFile, _, err = r.FormFile("image")
		} else {
			ErrorResponse(w, "Invalid upload type.", http.StatusBadRequest)
			return
		}

		if err != nil {
			ErrorResponse(w, "Error processing file!", http.StatusInternalServerError)
			return
		}

		defer uploadFile.Close()

		tmpFile, err := ioutil.TempFile(os.TempDir(), "image")
		if err != nil {
			ErrorResponse(w, "Unable to write to /tmp", http.StatusInternalServerError)
			return
		}

		defer tmpFile.Close()

		_, err = io.Copy(tmpFile, uploadFile)

		if err != nil {
			ErrorResponse(w, "Unable to copy image to disk!", http.StatusInternalServerError)
			return
		}

		// upload := FileUpload{
		// 	header.Filename,
		// 	os.TempDir() + tmpFile.Name(),
		// 	header.Header.Get("Content-Type"),
		// }

		resp := make(map[string]interface{})

		// TODO: Build JSON response

		js, err := json.Marshal(resp)
		if err != nil {
			ErrorResponse(w, "Unable to build JSON response!", http.StatusInternalServerError)
			return
		}

		w.Write(js)
	}

	http.HandleFunc("/upload", uploadHandler)

	http.ListenAndServe(":8080", nil)
}

func (s *Server) download(url string) (multipart.File, *multipart.FileHeader, error) {
	return nil, nil, errors.New("Not implemented")
}
