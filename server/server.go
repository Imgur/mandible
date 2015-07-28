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
}

type ServerResponse struct {
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Status  int         `json:"status"`
	Success *bool       `json:"success"` // the empty value is the nil pointer, because this is a computed property
}

func (resp *ServerResponse) Write(w http.ResponseWriter) {
	respBytes, _ := resp.json()
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

type UserError struct {
	UserFacingMessage error
	LogMessage        error
}

func NewServer(c *config.Configuration, strategy imageprocessor.ImageProcessorStrategy) *Server {
	factory := imagestore.NewFactory(c)
	httpclient := &http.Client{}
	stores := factory.NewImageStores()
	store := stores[0]

	hashGenerator := factory.NewHashGenerator(store)
	authenticator := &PassthroughAuthenticator{}
	return &Server{c, httpclient, store, hashGenerator, strategy, authenticator}
}

func NewAuthenticatedServer(c *config.Configuration, strategy imageprocessor.ImageProcessorStrategy, auth Authenticator) *Server {
	factory := imagestore.NewFactory(c)
	httpclient := &http.Client{}
	stores := factory.NewImageStores()
	store := stores[0]

	hashGenerator := factory.NewHashGenerator(store)
	return &Server{c, httpclient, store, hashGenerator, strategy, auth}
}

func (s *Server) uploadFile(uploadFile io.Reader, fileName string, thumbs []*uploadedfile.ThumbFile, user *AuthenticatedUser) ServerResponse {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "image")
	if err != nil {
		fmt.Println(err)

		return ServerResponse{
			Error:  "Unable to write to /tmp",
			Status: http.StatusInternalServerError,
		}
	}

	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, uploadFile)

	if err != nil {
		fmt.Println(err)

		return ServerResponse{
			Error:  "Unable to copy image to disk!",
			Status: http.StatusInternalServerError,
		}
	}

	upload, err := uploadedfile.NewUploadedFile(fileName, tmpFile.Name(), thumbs)
	defer upload.Clean()

	if err != nil {
		return ServerResponse{
			Error:  "Error detecting mime type!",
			Status: http.StatusInternalServerError,
		}
	}

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
	uploadFileFd, err := os.Open(uploadFilepath)

	if err != nil {
		log.Printf("Error opening processed output %+v at %s: %s", upload, uploadFilepath, err.Error())
		return ServerResponse{
			Error:  "Unable to save image!",
			Status: http.StatusInternalServerError,
		}
	}

	obj, err = s.ImageStore.Save(uploadFileFd, obj)
	if err != nil {
		log.Printf("Error saving processed output to store: %s", err.Error())
		return ServerResponse{
			Error:  "Unable to save image!",
			Status: http.StatusInternalServerError,
		}
	}

	thumbsResp := map[string]interface{}{}
	for _, t := range upload.GetThumbs() {
		thumbName := fmt.Sprintf("%s/%s", upload.GetHash(), t.GetName())
		tObj := factory.NewStoreObject(thumbName, upload.GetMime(), "t")

		tPath := t.GetPath()
		tFile, err := os.Open(tPath)

		if err != nil {
			return ServerResponse{
				Error:  "Unable to save thumbnail!",
				Status: http.StatusInternalServerError,
			}
		}

		tObj, err = s.ImageStore.Save(tFile, tObj)
		if err != nil {
			return ServerResponse{
				Error:  "Unable to save thumbnail!",
				Status: http.StatusInternalServerError,
			}
		}

		thumbsResp[t.GetName()] = tObj.Url
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

		return uploadFile, header.Filename, nil
	}

	var extractorUrl fileExtractor = func(r *http.Request) (uploadFile io.Reader, filename string, uerr *UserError) {
		url := r.FormValue("image")
		uploadFile, err := s.download(url)

		if err != nil {
			return nil, "", &UserError{LogMessage: err, UserFacingMessage: errors.New("Error downloading URL!")}
		}

		return uploadFile, path.Base(url), nil
	}

	var extractorBase64 fileExtractor = func(r *http.Request) (uploadFile io.Reader, filename string, uerr *UserError) {
		input := r.FormValue("image")
		b64data := input[strings.IndexByte(input, ',')+1:]

		uploadFile = base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64data))

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
				resp.Write(w)
				return
			}

			thumbs, err := parseThumbs(r)
			if err != nil {
				resp := ServerResponse{
					Status: http.StatusBadRequest,
					Error:  "Error parsing thumbnails!",
				}
				resp.Write(w)
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

			resp.Write(w)
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

	rootHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<html><head><title>An open source image uploader by Imgur</title></head><body style=\"background-color: #2b2b2b; color: white\">")
		fmt.Fprint(w, "Congratulations! Your image upload server is up and running. Head over to the <a style=\"color: #85bf25 \" href=\"https://github.com/Imgur/mandible\">github</a> page for documentation")
		fmt.Fprint(w, "<br/><br/><br/><img src=\"http://i.imgur.com/YbfUjs5.png?2\" />")
		fmt.Fprint(w, "</body></html>")
	}

	router := mux.NewRouter()

	router.HandleFunc("/file", uploadHandler(extractorFile, nil))
	router.HandleFunc("/url", uploadHandler(extractorUrl, nil))
	router.HandleFunc("/base64", uploadHandler(extractorBase64, nil))

	router.HandleFunc("/user/{user_id}/file", authenticatedEndpoint(uploadHandler, extractorBase64))
	router.HandleFunc("/user/{user_id}/url", authenticatedEndpoint(uploadHandler, extractorUrl))
	router.HandleFunc("/user/{user_id}/base64", authenticatedEndpoint(uploadHandler, extractorBase64))

	router.HandleFunc("/", rootHandler)

	muxer.Handle("/", router)
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
