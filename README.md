# ImgurGo

A ready-to-deploy uploader that you can run on AWS EC2 or Heroku. It accepts an image via a REST interface and returns information about a file. Also, supports processing steps such as compression and thumbnail generation.


[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy)

![TravisStatus](https://travis-ci.org/gophergala/ImgurGo.svg)

## Features
Supported file types
- JPG
- PNG
- GIF

Pluggable storage layers
- S3
- Local

Processing Steps:
- Compression
- Thumbnail generation

## REST API

Interfacing with ImgurGo is extremely simple:

### Upload an image file
`POST /file`

with the following multi-part/form-data
- ```image``` - file

### Upload an image from a URL
`POST /url`

with the following multi-part/form-data
- ```image``` - string

### Upload an image from base64 data
`POST /base64`

with the following multi-part/form-data
- ```image``` - image encoded as base64 data

### Example usage

*TODO*
