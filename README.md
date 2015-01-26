# ImgurGo

A ready-to-deploy uploader that you can run on AWS EC2 or Heroku. It accepts an image via a REST interface and returns information about a file. Also, supports processing steps such as compression and thumbnail generation.

![TravisStatus](https://travis-ci.org/gophergala/ImgurGo.svg)

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy)

## Features:
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

## Installation

### S3 Storage Layer

- Set the following environment variables
    - AWS_ACCESS_KEY_ID
    - AWS_SECRET_ACCESS_KEY
    - S3_BUCKET

## REST API:

Interfacing with ImgurGo is extremely simple:

### Upload an image file:
`POST /file`

with the following multi-part/form-data
- ```image``` - file

### Upload an image from a URL:
`POST /url`

with the following multi-part/form-data
- ```image``` - string

### Upload an image from base64 data:
`POST /base64`

with the following multi-part/form-data
- ```image``` - image encoded as base64 data

### Thumbnail Generation:

To generate thumbnails with the request, pass the following JSON as form-data, keyed under `thumbs`

```javascript
{
    "name1": {
        "width": x,
        "height": y,
        "shape": ("square" | "thumb" | "circle")
    },
    "name2": {
        "width": x2,
        "height": y2,
        "shape": ("square" | "thumb" | "circle")
    },

    ...
}
```

Note: Square thumbnails don't preserve aspect ratio, whereas the 'thumb' type does

## Example usage (assuming localhost)

### URL Upload with thumbnails:

```
curl -i http://127.0.0.1:8080/url \
-d 'image=http://i.imgur.com/s9zxmYe.jpg' \
-d 'thumbs={"small": {"width": 20, "height": 20, "shape": "square"}, "profile": {"width": 50, "height": 50, "shape": "circle"}}'
```
### Response:

```javascript
{
    "data": {
        "width": 380,
        "height": 430,
        "link": "https://s3.amazonaws.com/gophergala/original/CUqU4If",
        "mime": "image/jpeg",
        "name": "",
        "size": 82199,
        "thumbs": {
            "profile":"https://s3.amazonaws.com/gophergala/t/CUqU4If/profile",
            "small": "https://s3.amazonaws.com/gophergala/t/CUqU4If/small"
        }
    },
    "status": 200,
    "success": true
}
```
### File Upload with thumbnails:

```
curl -i http://127.0.0.1:8080/file \
-F 'image=@/tmp/cat.gif' \
-F 'thumbs={"small": {"width": 20, "height": 20, "shape": "square"}}'
```
### Response:

```javascript
{
    "data": {
        "width": 354,
        "height": 200,
        "link": "https://s3.amazonaws.com/gophergala/original/L4ASjMX",
        "mime": "image/gif",
        "name": "cat.gif",
        "size": 3511100,
        "thumbs": {
            "small":"https://s3.amazonaws.com/gophergala/t/L4ASjMX/small"
        }
    },
    "status": 200,
    "success": true
}
```
