package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/Imgur/mandible/config"
	"github.com/Imgur/mandible/imageprocessor"
	"github.com/Imgur/mandible/imagestore"
	"github.com/Imgur/mandible/uploadedfile"
)

type Server struct {
	Config            *config.Configuration
	HTTPClient        *http.Client
	ImageStore        imagestore.ImageStore
	hashGenerator     *imagestore.HashGenerator
	processorStrategy imageprocessor.ImageProcessorStrategy
	authenticator     Authenticator
	stats             RuntimeStats
}

type ServerResponse struct {
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Status  int         `json:"status"`
	Success *bool       `json:"success"` // the empty value is the nil pointer, because this is a computed property
}

func (resp *ServerResponse) Write(w http.ResponseWriter, s RuntimeStats) {
	respBytes, _ := resp.json()

	if resp.Status >= http.StatusBadRequest {
		log.Println(fmt.Sprintf("HTTP error: %d -- %s", resp.Status, resp.Error))
		s.Error(resp.Status)
	}

	w.WriteHeader(resp.Status)
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}

// The success property is a computed property on the response status
// This can't implement the MarshalJSON() interface sadly because it would be recursive
func (resp *ServerResponse) json() ([]byte, error) {
	var success bool
	success = (resp.Status == http.StatusOK)
	resp.Success = &success
	bytes, err := json.Marshal(resp)
	resp.Success = nil
	return bytes, err
}

type ImageResponse struct {
	Link    string                 `json:"link"`
	Mime    string                 `json:"mime"`
	Name    string                 `json:"name"`
	Hash    string                 `json:"hash"`
	Size    int64                  `json:"size"`
	Width   int                    `json:"width"`
	Height  int                    `json:"height"`
	OCRText string                 `json:"ocrtext"`
	Thumbs  map[string]interface{} `json:"thumbs"`
	UserID  string                 `json:"user_id"`
}

type OcrResponse struct {
	Hash    string `json:"hash"`
	OCRText string `json:"ocrtext"`
}

type UserError struct {
	UserFacingMessage error
	LogMessage        error
}

func NewServer(c *config.Configuration, strategy imageprocessor.ImageProcessorStrategy, stats RuntimeStats) *Server {
	factory := imagestore.NewFactory(c)
	httpclient := &http.Client{}
	stores := factory.NewImageStores()

	hashGenerator := factory.NewHashGenerator(stores)
	authenticator := &PassthroughAuthenticator{}
	return &Server{c, httpclient, stores, hashGenerator, strategy, authenticator, stats}
}

func NewAuthenticatedServer(c *config.Configuration, strategy imageprocessor.ImageProcessorStrategy, auth Authenticator, stats RuntimeStats) *Server {
	factory := imagestore.NewFactory(c)
	httpclient := &http.Client{}
	stores := factory.NewImageStores()

	hashGenerator := factory.NewHashGenerator(stores)
	return &Server{c, httpclient, stores, hashGenerator, strategy, auth, stats}
}

func (s *Server) uploadFile(uploadFile io.Reader, fileName string, thumbs []*uploadedfile.ThumbFile, user *AuthenticatedUser) ServerResponse {
	tmpFile, err := saveToTmp(uploadFile)
	if err != nil {
		return ServerResponse{
			Error:  "Error saving to disk!",
			Status: http.StatusInternalServerError,
		}
	}

	upload, err := uploadedfile.NewUploadedFile(fileName, tmpFile, thumbs)

	if err != nil {
		return ServerResponse{
			Error:  "Error detecting mime type!",
			Status: http.StatusInternalServerError,
		}
	}
	defer upload.Clean()

	processor, err := s.processorStrategy(s.Config, upload)
	if err != nil {
		log.Printf("Error creating processor factory: %s", err.Error())
		return ServerResponse{
			Error:  "Unable to process image!",
			Status: http.StatusInternalServerError,
		}
	}

	err = processor.Run(upload)
	if err != nil {
		log.Printf("Error processing %+v: %s", upload, err.Error())
		return ServerResponse{
			Error:  "Unable to process image!",
			Status: http.StatusInternalServerError,
		}
	}

	upload.SetHash(s.hashGenerator.Get())

	factory := imagestore.NewFactory(s.Config)
	obj := factory.NewStoreObject(upload.GetHash(), upload.GetMime(), "original")

	uploadFilepath := upload.GetPath()
	obj, err = s.ImageStore.Save(uploadFilepath, obj)
	if err != nil {
		log.Printf("Error saving processed output to store: %s", err.Error())
		return ServerResponse{
			Error:  "Unable to save image!",
			Status: http.StatusInternalServerError,
		}
	}

	thumbsResp, err := s.buildThumbResponse(upload)
	if err != nil {
		log.Printf("Error processing %+v: %s", upload, err.Error())
		return ServerResponse{
			Error:  "Unable to process thumbnail!",
			Status: http.StatusInternalServerError,
		}
	}

	size, err := upload.FileSize()
	if err != nil {
		return ServerResponse{
			Error:  "Unable to fetch image metadata!",
			Status: http.StatusInternalServerError,
		}
	}

	width, height, err := upload.Dimensions()

	if err != nil {
		return ServerResponse{
			Error:  "Error fetching upload dimensions: " + err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	var userID string
	if user != nil {
		userID = string(user.UserID)
	}

	resp := ImageResponse{
		Link:    obj.Url,
		Mime:    obj.MimeType,
		Hash:    upload.GetHash(),
		Name:    fileName,
		Size:    size,
		Width:   width,
		Height:  height,
		OCRText: upload.GetOCRText(),
		Thumbs:  thumbsResp,
		UserID:  userID,
	}

	return ServerResponse{
		Data:   resp,
		Status: http.StatusOK,
	}
}

type fileExtractor func(r *http.Request) (uploadFile io.Reader, filename string, uerr *UserError)

func (s *Server) Configure(muxer *http.ServeMux) {

	var extractorFile fileExtractor = func(r *http.Request) (uploadFile io.Reader, filename string, uerr *UserError) {
		uploadFile, header, err := r.FormFile("image")
		if err != nil {
			return nil, "", &UserError{LogMessage: err, UserFacingMessage: errors.New("Error processing file")}
		}

		s.stats.Upload("file")
		return uploadFile, header.Filename, nil
	}

	var extractorUrl fileExtractor = func(r *http.Request) (uploadFile io.Reader, filename string, uerr *UserError) {
		url := r.FormValue("image")
		uploadFile, err := s.download(url)

		if err != nil {
			return nil, "", &UserError{LogMessage: err, UserFacingMessage: errors.New("Error downloading URL!")}
		}

		s.stats.Upload("url")
		return uploadFile, path.Base(url), nil
	}

	var extractorBase64 fileExtractor = func(r *http.Request) (uploadFile io.Reader, filename string, uerr *UserError) {
		input := r.FormValue("image")
		b64data := input[strings.IndexByte(input, ',')+1:]

		uploadFile = base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64data))

		s.stats.Upload("base64")
		return uploadFile, "", nil
	}

	type uploadEndpoint func(fileExtractor, *AuthenticatedUser) http.HandlerFunc

	var uploadHandler uploadEndpoint = func(extractor fileExtractor, user *AuthenticatedUser) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			uploadFile, filename, uerr := extractor(r)
			if uerr != nil {
				log.Printf("Error extracting files: %s", uerr.LogMessage.Error())
				resp := ServerResponse{
					Status: http.StatusBadRequest,
					Error:  uerr.UserFacingMessage.Error(),
				}
				resp.Write(w, s.stats)
				return
			}

			thumbs, err := parseThumbs(r)
			if err != nil {
				resp := ServerResponse{
					Status: http.StatusBadRequest,
					Error:  "Error parsing thumbnails!",
				}
				resp.Write(w, s.stats)
				return
			}

			resp := s.uploadFile(uploadFile, filename, thumbs, user)

			switch uploadFile.(type) {
			case io.ReadCloser:
				defer uploadFile.(io.ReadCloser).Close()
				break
			default:
				break
			}

			resp.Write(w, s.stats)
		}
	}

	// Wrap an existing upload endpoint with authentication, returning a new endpoint that 4xxs unless authentication is passed.
	authenticatedEndpoint := func(endpoint uploadEndpoint, extractor fileExtractor) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			requestVars := mux.Vars(r)
			attemptedUserIdString, ok := requestVars["user_id"]

			// They didn't send a user ID to a /user endpoint
			if !ok || attemptedUserIdString == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			user, err := s.authenticator.GetUser(r)

			// Their HMAC was invalid or they are trying to upload to someone else's account
			if user == nil || err != nil || user.UserID != attemptedUserIdString {
				w.WriteHeader(http.StatusUnauthorized)
				log.Printf("Authentication error: %s", err.Error())
				return
			}

			handler := endpoint(extractor, user)
			handler(w, r)
		}
	}

	ocrHandler := func(w http.ResponseWriter, r *http.Request) {
		imageID := r.FormValue("uid")
		if imageID == "" {
			resp := ServerResponse{
				Status: http.StatusBadRequest,
				Error:  "Image ID must be passed as \"uid\"",
			}
			resp.Write(w, s.stats)
			return
		}

		factory := imagestore.NewFactory(s.Config)
		tObj := factory.NewStoreObject(imageID, "", "original")

		storeReader, err := s.ImageStore.Get(tObj)
		if err != nil {
			resp := ServerResponse{
				Status: http.StatusBadRequest,
				Error:  fmt.Sprintf("Error retrieving image with ID: %s", imageID),
			}
			resp.Write(w, s.stats)
			return
		}
		defer storeReader.Close()

		storeFile, err := saveToTmp(storeReader)
		if err != nil {
			resp := ServerResponse{
				Status: http.StatusBadRequest,
				Error:  fmt.Sprintf("Error saving original image to tmpfile: %s", imageID),
			}
			resp.Write(w, s.stats)
			return
		}
		defer os.Remove(storeFile)

		upload, err := uploadedfile.NewUploadedFile("", storeFile, nil)
		if err != nil {
			resp := ServerResponse{
				Error:  fmt.Sprintf("Unable to generate UploadedFile object: %s", imageID),
				Status: http.StatusInternalServerError,
			}
			resp.Write(w, s.stats)
			return
		}
		upload.SetHash(imageID)
		defer upload.Clean()

		//TODO: fix this sp error:
		processor := imageprocessor.DuelOCRStratagy()
		err = processor.Process(upload)
		if err != nil {
			log.Printf("Error runinng DuelOCRStrategy on %+v: %s", upload, err.Error())
			resp := ServerResponse{
				Error:  "Unable to execute OCR strategy",
				Status: http.StatusInternalServerError,
			}
			resp.Write(w, s.stats)
			return
		}

		ocrResp := OcrResponse{
			Hash:    upload.GetHash(),
			OCRText: upload.GetOCRText(),
		}

		resp := ServerResponse{
			Data:   ocrResp,
			Status: http.StatusOK,
		}

		resp.Write(w, s.stats)
	}

	thumbnailHandler := func(w http.ResponseWriter, r *http.Request) {
		imageID := r.FormValue("uid")

		factory := imagestore.NewFactory(s.Config)
		tObj := factory.NewStoreObject(imageID, "", "original")

		thumbs, err := parseThumbs(r)
		if err != nil {
			resp := ServerResponse{
				Status: http.StatusBadRequest,
				Error:  "Error parsing thumbnails!",
			}
			resp.Write(w, s.stats)
			return
		}

		if len(thumbs) != 1 {
			resp := ServerResponse{
				Status: http.StatusBadRequest,
				Error:  "Wrong number of thumbnails, expected 1",
			}
			resp.Write(w, s.stats)
			return
		}

		storeReader, err := s.ImageStore.Get(tObj)
		if err != nil {
			resp := ServerResponse{
				Status: http.StatusNotFound,
				Error:  fmt.Sprintf("Error retrieving image with ID: %s", imageID),
			}
			resp.Write(w, s.stats)
			return
		}
		defer storeReader.Close()

		storeFile, err := saveToTmp(storeReader)
		if err != nil {
			resp := ServerResponse{
				Status: http.StatusInternalServerError,
				Error:  "Error saving original Image!",
			}
			resp.Write(w, s.stats)
			return
		}
		defer os.Remove(storeFile)

		upload, err := uploadedfile.NewUploadedFile("", storeFile, thumbs)
		if err != nil {
			log.Printf("Error processing %+v: %s", storeFile, err.Error())
			resp := ServerResponse{
				Error:  "Unable to process thumbnail!",
				Status: http.StatusInternalServerError,
			}
			resp.Write(w, s.stats)
			return
		}
		upload.SetHash(imageID)
		defer upload.Clean()

		processor, _ := imageprocessor.ThumbnailStrategy(s.Config, upload)
		err = processor.Run(upload)
		if err != nil {
			log.Printf("Error processing %+v: %s", upload, err.Error())
			resp := ServerResponse{
				Error:  "Unable to process thumbnail!",
				Status: http.StatusInternalServerError,
			}
			resp.Write(w, s.stats)
			return
		}

		ts := upload.GetThumbs()
		t := ts[0]

		if !t.GetNoStore() {
			thumbName := fmt.Sprintf("%s/%s", upload.GetHash(), t.Name)
			tObj = factory.NewStoreObject(thumbName, upload.GetMime(), "thumbnail")
			err = tObj.Store(t, s.ImageStore)
			if err != nil {
				log.Printf("Error storing %+v: %s", t, err.Error())
				resp := ServerResponse{
					Error:  "Unable to store thumbnail!",
					Status: http.StatusInternalServerError,
				}
				resp.Write(w, s.stats)
				return
			}
		}

		s.stats.Thumbnail(t.Name)

		http.ServeFile(w, r, t.GetPath())
	}

	rootHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<html><head><title>An open source image uploader by Imgur</title></head><body style=\"background-color: #2b2b2b; color: white\">")
		fmt.Fprint(w, "Congratulations! Your image upload server is up and running. Head over to the <a style=\"color: #85bf25 \" href=\"https://github.com/Imgur/mandible\">github</a> page for documentation")
		fmt.Fprint(w, "<br/><br/><br/><img src=\"http://i.imgur.com/YbfUjs5.png?2\" />")
		fmt.Fprint(w, "</body></html>")
	}

	requestMiddleware := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			s.stats.Request(r.URL.Path)

			if os.Getenv("MANDIBLE_DEBUG") == "true" {
				r.ParseForm()
				log.Printf("Request url: %s with get params: %v and Headers: %v", r.URL.Path, r.Form, r.Header)
			}

			start := time.Now()
			handler(w, r)
			elapsed := time.Since(start)

			s.stats.ResponseTime(elapsed, r.URL.Path)
		}
	}

	router := mux.NewRouter()

	router.HandleFunc("/file", requestMiddleware(uploadHandler(extractorFile, nil)))
	router.HandleFunc("/url", requestMiddleware(uploadHandler(extractorUrl, nil)))
	router.HandleFunc("/base64", requestMiddleware(uploadHandler(extractorBase64, nil)))

	router.HandleFunc("/user/{user_id}/file", requestMiddleware(authenticatedEndpoint(uploadHandler, extractorBase64)))
	router.HandleFunc("/user/{user_id}/url", requestMiddleware(authenticatedEndpoint(uploadHandler, extractorUrl)))
	router.HandleFunc("/user/{user_id}/base64", requestMiddleware(authenticatedEndpoint(uploadHandler, extractorBase64)))

	router.HandleFunc("/thumbnail", requestMiddleware(thumbnailHandler))

	router.HandleFunc("/ocr", requestMiddleware(ocrHandler))
	router.HandleFunc("/", requestMiddleware(rootHandler))

	muxer.Handle("/", router)
}

func (s *Server) buildThumbResponse(upload *uploadedfile.UploadedFile) (map[string]interface{}, error) {
	factory := imagestore.NewFactory(s.Config)
	thumbsResp := map[string]interface{}{}

	for _, t := range upload.GetThumbs() {
		thumbName := fmt.Sprintf("%s/%s", upload.GetHash(), t.Name)
		tObj := factory.NewStoreObject(thumbName, upload.GetMime(), "thumbnail")
		err := tObj.Store(t, s.ImageStore)
		if err != nil {
			return nil, err
		}

		s.stats.Thumbnail(t.Name)
		thumbsResp[t.Name] = tObj.Url
	}

	return thumbsResp, nil
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

	type ThumbRequest struct {
		Width         int    `json:"width"`
		MaxWidth      int    `json:"max_width"`
		Height        int    `json:"height"`
		MaxHeight     int    `json:"max_height"`
		Shape         string `json:"shape"`
		CropGravity   string `json:"crop_gravity"`
		CropHeight    int    `json:"crop_height"`
		CropWidth     int    `json:"crop_width"`
		Quality       int    `json:"quality"`
		CropRatio     string `json:"crop_ratio"`
		DesiredFormat string `json:"format"`
		NoStore       bool   `json:"nostore"`
	}
	var thumbRequests map[string]ThumbRequest
	err := json.Unmarshal([]byte(thumbString), &thumbRequests)
	if err != nil {
		fmt.Println(err.Error())
		return nil, errors.New("Error parsing thumbnail JSON!")
	}

	var thumbs []*uploadedfile.ThumbFile
	for name, thumbRequest := range thumbRequests {
		thumb := uploadedfile.NewThumbFile(
			thumbRequest.Width,
			thumbRequest.MaxWidth,
			thumbRequest.Height,
			thumbRequest.MaxHeight,
			name,
			thumbRequest.Shape,
			"", // shape
			thumbRequest.CropGravity,
			thumbRequest.CropWidth,
			thumbRequest.CropHeight,
			thumbRequest.CropRatio,
			thumbRequest.Quality,
			thumbRequest.DesiredFormat,
			thumbRequest.NoStore,
		)

		thumbs = append(thumbs, thumb)
	}

	return thumbs, nil
}

func saveToTmp(upload io.Reader) (string, error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "image")
	if err != nil {
		fmt.Println(err)

		return "", err
	}

	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, upload)
	if err != nil {
		fmt.Println(err)

		return "", err
	}

	return tmpFile.Name(), nil
}
