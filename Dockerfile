FROM golang:1.3.3

RUN apt-get update && apt-get install -y --force-yes libjpeg-progs optipng imagemagick gifsicle exiftool aspell aspell-en libaspell-dev tesseract-ocr tesseract-ocr-eng
COPY docker/meme.traineddata /usr/share/tesseract-ocr/tessdata/meme.traineddata

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

CMD ["go-wrapper", "run"]

RUN mkdir -p /etc/mandible/
COPY docker/conf.json /etc/mandible/

COPY . /go/src/app
RUN go-wrapper download
RUN go-wrapper install
