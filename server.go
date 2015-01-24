package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type FileUpload struct {
	filename string
	path     string
	mime     string
}

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
	http.ListenAndServe(":8080", nil)

	uploadHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		uploadType := r.FormValue("type")

		switch uploadType {
		case "base64":
		case "url":
		case "file":
			uploadFile, _, err := r.FormFile("image")
			if err != nil {
				ErrorResponse(w, "Error processing form file!", http.StatusInternalServerError)
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

		// TODO: Pass `upload` over channel for processing
		default:
			ErrorResponse(w, "Invalid type!", http.StatusBadRequest)
			return
		}

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
}
