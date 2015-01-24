# ImgurGo

A ready-to-deploy uploader that you can run on AWS EC2 or Heroku. It accepts an image via a REST interface and returns information about a file. Also, supports processing steps such as compression and thumbnail generation.

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
```POST /upload```

with the following multi-part/form-data

- `type` - string
    - Value should be one of:
        - `url`
        - `file`
        - `base64`
- ```image``` - mixed
    - string, if `type` is `url`
    - base64 data, if `type` is `base64`
    - image file, if `type` is `file`

### Example usage

*TODO*
