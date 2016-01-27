# mandible ![TravisStatus](https://travis-ci.org/Imgur/mandible.svg) [![Coverage Status](https://coveralls.io/repos/Imgur/mandible/badge.svg)](https://coveralls.io/r/Imgur/mandible)

A ready-to-deploy uploader that you can run on AWS EC2 or Heroku. It accepts an image via a REST interface and returns information about a file. Also, supports processing steps such as compression and thumbnail generation.


[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy)


## Features:
Supported file types
- JPG
- PNG
- GIF

Pluggable storage layers
- S3
- Local

Pluggable authentication scheme
- Time-grant HMAC

Processing Steps:
- Compression
- Thumbnail generation

## Installation

### Docker

Pull down the mandible config file and edit it: 

```
wget https://raw.githubusercontent.com/Imgur/mandible/master/config/default.conf.json -O ~/mandible/conf.json
```
```
vim ~/mandible/conf.json
```

To start mandible (port settings could change based on your conf.json):

```
docker run --name mandible -v ~/mandible:/etc/mandible -d -p 80:8080 imgur/mandible
```

To stop mandible:

```
docker stop mandible
```

To run it again:

```
docker run mandible
```
### (Optional) Authentication

- Set the following environment variable
    - AUTHENTICATION_HMAC_KEY

### S3 Storage Layer
Add the following to the `Stores` array in your conf.json file:

```
    {
        "Type" : "s3",
        "BucketName" : "",
        "AWSKey": "",
        "AWSSecret": "",
        "StoreRoot" : "",
        "Region" : "us-east-1",
        "NamePathRegex" : "",
        "NamePathMap" : "${ImageSize}/${ImageName}"
    }
```


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

### Authenticated upload

Uses HTTP headers `Authentication` and `X-Authentication-HMAC`. Generate HMACs by base64-encoding a JSON blob like below. [Example MAC generator](http://play.golang.org/p/3otGr8LBZt). 
Supplying the client with the Authentication blob and MAC is out of scope for this project. In the future we will support symmetric and asymmetric encryption of the authentication blobs.

#### Request to my own account with proper authorization:

```
curl -i http://127.0.0.1:8080/user/1/url \
-d 'image=http://i.imgur.com/s9zxmYe.jpg' \
-H 'Authorization: {"user_id":1,"grant_time":"2010-06-01T00:00:00Z","grant_duration_sec":31536000}' \
-H 'X-Authorization-HMAC: tCtGb04n4nvd/94+Xd6vAx9+pJw51ZmX1vH7E+BlTtc='
```

#### Response:
```javascript
{"data":{"link":"/tmp/original/J/a/Jafq9IH","mime":"image/jpeg","name":"s9zxmYe.jpg","hash":"Jafq9IH","size":81881,"width":380,"height":430,"ocrtext":"change\np.roject .\n\n  \n  \n\n  forg@ot to git p.ull before\n- .-+#~+):,-r,ad)q..,i,ng so/ /","thumbs":{},"user_id":"\u0001"},"status":200,"success":true}
```

#### Request to other user's account:

```
curl -i http://127.0.0.1:8080/user/2/url \
-d 'image=http://i.imgur.com/s9zxmYe.jpg' \
-H 'Authorization: {"user_id":1,"grant_time":"2010-06-01T00:00:00Z","grant_duration_sec":31536000}' \
-H 'X-Authorization-HMAC: tCtGb04n4nvd/94+Xd6vAx9+pJw51ZmX1vH7E+BlTtc='
```

#### Response:
```
HTTP/1.1 401 Unauthorized
Date: Mon, 08 Jun 2015 21:04:41 GMT
Content-Length: 0
Content-Type: text/plain; charset=utf-8
```


#### HMAC prevents account forgery

```
curl -i http://127.0.0.1:8080/user/1/url \
-d 'image=http://i.imgur.com/s9zxmYe.jpg' \
-H 'Authorization: {"user_id":1,"grant_time":"2010-06-01T00:00:00Z","grant_duration_sec":31536000}' \
-H 'X-Authorization-HMAC: foobar'
```

#### Response:
```javascript
HTTP/1.1 401 Unauthorized
Date: Mon, 08 Jun 2015 21:04:41 GMT
Content-Length: 0
Content-Type: text/plain; charset=utf-8
```

## Contributing

The easiest way to develop on this project is to use the built-in docker image. We are using the Go 1.5 vendor experiment, which means if
you import a package you must vendor the source code into this repository using Godep. 
