package storage

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	fileLastModifiedLayout = "2006-01-02T15:04:05.999999"
	queryFormat            = "format"
	queryJSON              = "json"
	headMethod             = "HEAD"
	getMethod              = "GET"
	postMethod             = "POST"
	putMethod              = "PUT"
	deleteMethod           = "DELETE"
	authTokenHeader        = "X-Auth-Token"
	objectCountHeader      = "X-Account-Object-Count"
	bytesUsedHeader        = "X-Account-Bytes-Used"
	containerCountHeader   = "X-Account-Container-Count"
	recievedBytesHeader    = "X-Received-Bytes"
	transferedBytesHeader  = "X-Transfered-Bytes"
	uint64BitSize          = 64
	uint64Base             = 10
	// EnvUser is environmental variable for selectel api username
	EnvUser = "SELECTEL_USER"
	// EnvKey is environmental variable for selectel api key
	EnvKey = "SELECTEL_KEY"
)

var (
	// ErrorObjectNotFound occurs when server returns 404
	ErrorObjectNotFound = errors.New("Object not found")
	// ErrorBadResponce occurs when server returns unexpected code
	ErrorBadResponce = errors.New("Unable to process api responce")
	// ErrorBadName
	ErrorBadName = errors.New("Bad container/object name provided")
	// ErrorBadJSON occurs on unmarhalling error
	ErrorBadJSON = errors.New("Unable to parse api responce")
)

// Client is selectel storage api client
type Client struct {
	storageURL  *url.URL
	token       string
	tokenExpire int
	expireFrom  *time.Time
	user        string
	key         string
	client      DoClient
	file        fileMock
	debug       bool
}

type ClientCredentials struct {
	Token      string
	Debug      bool
	Expire     int
	ExpireFrom *time.Time
	URL        string
}

func NewFromCache(data []byte) (API, error) {
	var (
		cache = new(ClientCredentials)
		err   error
	)
	decorer := gob.NewDecoder(bytes.NewBuffer(data))
	if err = decorer.Decode(cache); err != nil {
		return nil, err
	}
	c := newClient(new(http.Client))
	c.token = cache.Token
	c.tokenExpire = cache.Expire
	c.debug = cache.Debug
	c.expireFrom = cache.ExpireFrom
	c.storageURL, err = url.Parse(cache.URL)
	if err != nil {
		return nil, ErrorBadCredentials
	}
	return c, nil
}

func (c *Client) Credentials() (cache ClientCredentials) {
	cache.URL = c.storageURL.String()
	cache.Expire = c.tokenExpire
	cache.ExpireFrom = c.expireFrom
	cache.Token = c.token
	cache.Debug = c.debug

	return cache
}

func (c *Client) Dump() ([]byte, error) {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	if err := encoder.Encode(c.Credentials()); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// StorageInformation contains some usefull metrics about storage for current user
type StorageInformation struct {
	ObjectCount     uint64
	BytesUsed       uint64
	ContainerCount  uint64
	RecievedBytes   uint64
	TransferedBytes uint64
}

// API for selectel storage
type API interface {
	DoClient
	Info() StorageInformation
	Upload(reader io.Reader, container, filename, t string) error
	UploadFile(filename, container string) error
	Auth(user, key string) error
	Debug(debug bool)
	Token() string
	C(string) ContainerAPI
	Container(string) ContainerAPI
	RemoveObject(container, filename string) error
	URL(container, filename string) string
	CreateContainer(name string, private bool) (ContainerAPI, error)
	RemoveContainer(name string) error
	// ObjectInfo returns information about object in container
	ObjectInfo(container, filename string) (f ObjectInfo, err error)
	ObjectsInfo(container string) ([]ObjectInfo, error)
	ContainerInfo(name string) (info ContainerInfo, err error)
	ContainersInfo() ([]ContainerInfo, error)
	Containers() ([]ContainerAPI, error)
	Credentials() (cache ClientCredentials)
	Dump() ([]byte, error)
}

// DoClient is mock of http.Client
type DoClient interface {
	Do(request *http.Request) (*http.Response, error)
}

// setClient sets client
func (c *Client) setClient(client DoClient) {
	c.client = client
}

func (c *Client) Debug(debug bool) {
	c.debug = debug
}

// ContainersInfo return all container-specific information from storage
func (c *Client) ContainersInfo() ([]ContainerInfo, error) {
	info := []ContainerInfo{}
	request, err := c.NewRequest(getMethod, nil)
	if err != nil {
		return nil, err
	}
	query := request.URL.Query()
	query.Add(queryFormat, queryJSON)
	request.URL.RawQuery = query.Encode()
	res, err := c.Do(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, ErrorBadResponce
	}
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&info); err != nil {
		return nil, ErrorBadJSON
	}
	return info, nil
}

// Containers return all containers from storage
func (c *Client) Containers() ([]ContainerAPI, error) {
	info, err := c.ContainersInfo()
	if err != nil {
		return nil, err
	}
	containers := []ContainerAPI{}
	for _, container := range info {
		containers = append(containers, c.Container(container.Name))
	}
	return containers, nil
}

// ObjectsInfo returns information about all objects in container
func (c *Client) ObjectsInfo(container string) ([]ObjectInfo, error) {
	info := []ObjectInfo{}
	request, err := c.NewRequest(getMethod, nil, container)
	if err != nil {
		return nil, err
	}
	query := request.URL.Query()
	query.Add(queryFormat, queryJSON)
	request.URL.RawQuery = query.Encode()
	res, err := c.Do(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, ErrorObjectNotFound
	}
	if res.StatusCode != http.StatusOK {
		return nil, ErrorBadResponce
	}
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&info); err != nil {
		return nil, ErrorBadJSON
	}
	for i, v := range info {
		info[i].LastModified, err = time.Parse(fileLastModifiedLayout, v.LastModifiedStr)
		if err != nil {
			return info, err
		}
	}
	return info, nil
}

// DeleteObject removes object from specified container
func (c *Client) RemoveObject(container, filename string) error {
	request, err := c.NewRequest(deleteMethod, nil, container, filename)
	if err != nil {
		return err
	}
	res, err := c.Do(request)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusNotFound {
		return ErrorObjectNotFound
	}
	if res.StatusCode == http.StatusNoContent {
		return nil
	}
	return ErrorBadResponce
}

// Info returns StorageInformation for current user
func (c *Client) Info() (info StorageInformation) {
	request, err := c.NewRequest(getMethod, nil)
	if err != nil {
		return
	}
	res, err := c.do(request)
	if err != nil {
		return
	}
	parse := func(key string) uint64 {
		v, _ := strconv.ParseUint(res.Header.Get(key), uint64Base, uint64BitSize)
		return v
	}
	info.BytesUsed = parse(bytesUsedHeader)
	info.ObjectCount = parse(objectCountHeader)
	info.ContainerCount = parse(containerCountHeader)
	info.RecievedBytes = parse(recievedBytesHeader)
	info.TransferedBytes = parse(transferedBytesHeader)
	return
}

// URL returns url for file in container
func (c *Client) URL(container, filename string) string {
	return c.url(container, filename)
}

// Do performs request with auth token
func (c *Client) Do(request *http.Request) (res *http.Response, err error) {
	return c.do(request)
}

func (c *Client) do(request *http.Request) (res *http.Response, err error) {
	// prevent null pointer dereference
	if request.Header == nil {
		request.Header = http.Header{}
	}
	// check for token expiration / first request with async auth
	if request.URL.String() != authURL && c.Expired() {
		log.Println("[selectel]", "token expired, performing auth")
		if err = c.Auth(c.user, c.key); err != nil {
			return
		}
		// fix hostname of request
		c.fixURL(request)
	}
	// add auth token to headers
	if !blank(c.token) {
		request.Header.Add(authTokenHeader, c.token)
	}
	if c.debug {
		// perform request and record time elapsed
		start := time.Now().Truncate(time.Millisecond)
		res, err = c.client.Do(request)
		stop := time.Now().Truncate(time.Millisecond)
		duration := stop.Sub(start)
		// log error
		if err != nil {
			log.Println(request.Method, request.URL.String(), err, duration)
			return
		}
		// log request
		log.Println(request.Method, request.URL.String(), res.StatusCode, duration)
		// check for auth code
	} else {
		res, err = c.client.Do(request)
		if err != nil {
			return
		}
	}
	if res.StatusCode == http.StatusUnauthorized {
		c.expireFrom = nil // ensure that next request will force authentication
		return nil, ErrorAuth
	}
	return
}

func (c *Client) NewRequest(method string, body io.Reader, parms ...string) (*http.Request, error) {
	var badName bool
	for i := range parms {
		// check for length
		if len(parms[i]) > 256 {
			badName = true
		}
		// todo: check for trialing slash
		parms[i] = url.QueryEscape(parms[i])
	}
	req, err := http.NewRequest(method, c.url(parms...), body)
	if err != nil || badName {
		return nil, ErrorBadName
	}
	return req, nil
}

func (c *Client) fixURL(request *http.Request) error {
	newRequest, err := http.NewRequest(request.Method, c.url(request.URL.Path), request.Body)
	*request = *newRequest
	return err
}

func (c *Client) url(postfix ...string) string {
	path := strings.Join(postfix, "/")
	if c.storageURL == nil {
		return path
	}
	return fmt.Sprintf("%s%s", c.storageURL, path)
}

// New returns new selectel storage api client
func New(user, key string) (API, error) {
	client := newClient(new(http.Client))
	return client, client.Auth(user, key)
}

// NewAsync returns new api client and lazily performs auth
func NewAsync(user, key string) API {
	c := newClient(new(http.Client))
	if blank(user) || blank(key) {
		panic(ErrorBadCredentials)
	}
	c.user = user
	c.key = key
	return c
}

func newClient(client *http.Client) *Client {
	c := new(Client)
	c.client = client
	return c
}

// NewEnv acts as New, but reads credentials from environment
func NewEnv() (API, error) {
	user := os.Getenv(EnvUser)
	key := os.Getenv(EnvKey)
	return New(user, key)
}

func blank(s string) bool {
	return len(s) == 0
}
