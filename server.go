package main

import (
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

type Server struct {
	Config *Configuration
}

func CreateServer(c *Configuration) *Server {
	return &Server{c}
}

func (s *Server) initServer() {
	fileHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		uploadFile, _, err := r.FormFile("image")

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

		// TODO: Build JSON respons

		Response(w, resp)
	}

	urlHandler := func(w http.ResponseWriter, r *http.Request) {
		_, _, err := s.download(r.FormValue("image"))

		if err != nil {
			ErrorResponse(w, "Error dowloading URL!", http.StatusInternalServerError)
			return
		}
	}

	http.HandleFunc("/file", fileHandler)
	http.HandleFunc("/url", urlHandler)

	http.ListenAndServe(":8080", nil)
}

func (s *Server) download(url string) (multipart.File, *multipart.FileHeader, error) {
	return nil, nil, errors.New("Not implemented")
}
