package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/ernado/selectel/storage"
	"github.com/jwaldrip/odin/cli"
	"github.com/olekukonko/tablewriter"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

const (
	envKey        = storage.EnvKey
	envUser       = storage.EnvUser
	version       = "1.1"
	cacheFilename = "~selct.cache~" + version
	envCache      = "SELECTEL_CACHE"
	envContainer  = "SELECTEL_CONTAINER"
)

var (
	client         = cli.New(version, "Selectel storage command line client", connect)
	user, key      string
	container      string
	api            storage.API
	debug          bool
	cache          bool
	cacheSecure    bool
	errorNotEnough = errors.New("Not enought arguments")
)

func encryptionKey() []byte {
	hasher := sha256.New()
	hasher.Write([]byte("selectel storage command line client"))
	hasher.Write([]byte(key))
	hasher.Write([]byte(user))
	return hasher.Sum(nil)
}

func encrypt(data []byte) []byte {
	block, err := aes.NewCipher(encryptionKey())
	if err != nil {
		panic(err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], data)
	return ciphertext
}

func decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(encryptionKey())
	if err != nil {
		return nil, err
	}

	if len(data) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(data, data)

	return data, nil
}

func init() {
	client.DefineBoolFlagVar(&debug, "debug", false, "debug mode")
	client.DefineBoolFlagVar(&cache, "cache", false, fmt.Sprintf("cache token in file (%s)", envCache))
	client.DefineBoolFlagVar(&cacheSecure, "cache.secure", true, "encrypt/decrypt token with user-key pair (true by default)")
	client.DefineStringFlag("key", "", fmt.Sprintf("selectel storage key (%s)", envKey))
	client.AliasFlag('k', "key")
	client.DefineStringFlag("user", "", fmt.Sprintf("selectel storage user (%s)", envUser))
	client.AliasFlag('u', "user")
	client.DefineStringFlag("container", "", fmt.Sprintf("default container (%s)", envContainer))
	client.AliasFlag('c', "container")

	infoCommand := client.DefineSubCommand("info", "print information about storage/container/object", wrap(info))
	infoCommand.DefineStringFlag("type", "storage", "storage, container or object")
	infoCommand.AliasFlag('t', "type")

	listCommand := client.DefineSubCommand("list", "list objects in container/storage", wrap(list))
	listCommand.DefineStringFlag("type", "storage", "storage or container")
	listCommand.AliasFlag('t', "type")

	client.DefineSubCommand("upload", "upload object to container", wrap(upload))
	downloadCommand := client.DefineSubCommand("download", "download object from container", wrap(download))
	downloadCommand.DefineStringFlag("path", "", "destination path")
	downloadCommand.AliasFlag('p', "path")

	client.DefineSubCommand("create", "create container", wrap(create))

	removeCommand := client.DefineSubCommand("remove", "remove object or container", wrap(remove))
	removeCommand.DefineStringFlag("type", "object", "container or object")
	removeCommand.DefineBoolFlag("force", false, "remove container with files")
	removeCommand.AliasFlag('f', "force")
	removeCommand.AliasFlag('t', "type")
}

func readFlag(c cli.Command, name, env string) string {
	if len(os.Getenv(env)) > 0 {
		return os.Getenv(env)
	}
	return c.Flag(name).String()
}

func blank(s string) bool {
	return len(s) == 0
}

func load() ([]byte, error) {
	f, err := os.Open(cacheFilename)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if !cacheSecure {
		return data, nil
	}
	return decrypt(data)
}

// connect reads credentials and performs auth
func connect(c cli.Command) {
	var err error

	key = readFlag(c, "key", envKey)
	user = readFlag(c, "user", envUser)
	container = readFlag(c, "container", envContainer)

	if strings.ToLower(os.Getenv(envCache)) == "true" {
		cache = true
	}

	if cache {
		var data []byte
		data, err = load()
		if err != nil {
			log.Println(err)
		} else {
			api, err = storage.NewFromCache(data)
			if err == nil {
				return
			} else {
				log.Println("unable to load from cache:", err)
			}
		}
	} else {
		os.Remove(cacheFilename)
	}

	// checking for blank credentials
	if blank(key) || blank(user) && api != nil {
		log.Fatal(storage.ErrorBadCredentials)
	}

	// connencting to api
	api = storage.NewAsync(user, key)
	api.Debug(debug)
	if err = api.Auth(user, key); err != nil {
		log.Fatal(err)
	}
}

func wrap(callback func(cli.Command)) func(cli.Command) {
	return func(c cli.Command) {
		connect(c.Parent())
		defer func() {
			if !cache {
				return
			}
			data, _ := api.Dump()
			if cacheSecure {
				data = encrypt(data)
			}
			f, _ := os.Create(cacheFilename)
			f.Write(data)
		}()
		callback(c)
	}
}

// info prints information about storage
func info(c cli.Command) {
	var (
		containerName = container
		objectName    string
		data          interface{}
		err           error
		arglen        = len(c.Args())
		command       = c.Flag("type").String()
	)

	defer func() {
		if err != nil {
			log.Fatal(err)
		}
		if blank(containerName) || command == "storage" {
			data = api.Info()
		} else {
			containerApi := api.Container(containerName)
			if blank(objectName) {
				data, err = containerApi.Info()
			} else {
				data, err = containerApi.Object(objectName).Info()
			}
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", data)
	}()

	if arglen > 0 {
		if command == "container" {
			containerName = c.Arg(0).String()
			return
		}
		command = "object"
		if !blank(containerName) && arglen == 1 {
			objectName = c.Arg(0).String()
			return
		}
		if arglen == 2 {
			containerName = c.Arg(0).String()
			objectName = c.Arg(1).String()
			return
		}
	}
	if command == "container" && !blank(containerName) {
		return
	}
	if command == "storage" {
		return
	}
	err = errorNotEnough
}

func remove(c cli.Command) {
	var (
		arglen  = len(c.Args())
		object  string
		err     error
		message string
		objects []storage.ObjectAPI
	)
	if arglen == 2 {
		container = c.Arg(0).String()
		object = c.Arg(1).String()
	}
	if arglen == 1 {
		if c.Flag("type").String() == "container" {
			container = c.Arg(0).String()
		} else {
			object = c.Arg(0).String()
		}
	}
	if blank(container) {
		log.Fatal(errorNotEnough)
	}
	if blank(object) {
		containerApi := api.Container(container)
		err = containerApi.Remove()

		// forced removal of container
		if err == storage.ErrorConianerNotEmpty && c.Flag("force").Get().(bool) {
			fmt.Println("removing all objects of", container)
			objects, err = containerApi.Objects()
			if err != nil {
				log.Fatal(err)
			}
			for _, object := range objects {
				err = object.Remove()
				// skipping NotFound errors as non-critical
				if err != nil && err != storage.ErrorObjectNotFound {
					log.Fatal(err)
				}
			}
			err = containerApi.Remove()
		}
		message = fmt.Sprintf("container %s removed", container)
	} else {
		err = api.Container(container).Object(object).Remove()
		message = fmt.Sprintf("object %s removed in container %s", object, container)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(message)
}

func create(c cli.Command) {
	if len(c.Args()) == 0 {
		log.Fatal(errorNotEnough)
	}
	var name = c.Arg(0).String()
	if _, err := api.CreateContainer(name, false); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created container %s\n", name)
}

func upload(c cli.Command) {
	var path string
	switch len(c.Args()) {
	case 1:
		path = c.Arg(0).String()
	case 2:
		container = c.Arg(0).String()
		path = c.Arg(1).String()
	}
	if blank(container) || blank(path) {
		log.Fatal(errorNotEnough)
	}

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	stat, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	ext := filepath.Ext(path)
	mimetype := mime.TypeByExtension(ext)
	bar := pb.New64(stat.Size()).SetUnits(pb.U_BYTES)
	bar.Start()
	reader := io.TeeReader(f, bar)
	if err := api.Container(container).Upload(reader, stat.Name(), mimetype); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("uploaded to %s\n", container)
}

func list(c cli.Command) {
	var (
		arglen = len(c.Args())
		table  = tablewriter.NewWriter(os.Stdout)
	)
	if arglen == 0 && (blank(container) || c.Flag("type").String() == "storage") {
		containers, err := api.ContainersInfo()
		if err != nil {
			log.Fatal(err)
		}
		table.SetHeader([]string{"Name", "Objects", "Type"})
		for _, cont := range containers {
			v := []string{cont.Name, fmt.Sprint(cont.ObjectCount), cont.Type}
			table.Append(v)
		}
		table.Render()
		return
	}
	if arglen == 1 {
		container = c.Arg(0).String()
	}
	if blank(container) {
		log.Fatal(errorNotEnough)
	}
	objects, err := api.Container(container).ObjectsInfo()
	if err != nil {
		log.Fatal(err)
	}
	table.SetHeader([]string{"Name", "Size", "Downloaded"})
	for _, object := range objects {
		v := []string{object.Name, fmt.Sprint(object.Size), fmt.Sprint(object.Downloaded)}
		table.Append(v)
	}
	table.Render()
}

func download(c cli.Command) {
	var (
		arglen     = len(c.Args())
		objectName string
		path       = c.Flag("path").String()
	)
	switch arglen {
	case 1:
		objectName = c.Arg(0).String()
	case 2:
		objectName = c.Arg(1).String()
		container = c.Arg(0).String()
	}
	if blank(container) || blank(objectName) {
		log.Fatal(errorNotEnough)
	}
	if blank(path) {
		path = objectName
	}
	reader, err := api.Container(container).Object(objectName).GetReader()
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	fmt.Printf("downloading %s->%s from %s\n", objectName, path, container)
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	n, err := io.Copy(f, reader)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("downloaded %s, %d bytes\n", objectName, n)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered", r)
		}
	}()
	client.Start()
}
