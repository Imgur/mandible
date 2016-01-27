FROM golang:1.5
RUN apt-get update && apt-get install -yqq aspell aspell-en libaspell-dev tesseract-ocr tesseract-ocr-eng imagemagick optipng exiftool libjpeg-progs webp
ADD docker/meme.traineddata /usr/share/tesseract-ocr/tessdata/meme.traineddata
RUN mkdir -p /etc/mandible /tmp/imagestore
ENV MANDIBLE_CONF /etc/mandible/conf.json
ENV GO15VENDOREXPERIMENT 1
ADD . /go/src/github.com/Imgur/mandible
WORKDIR /go/src/github.com/Imgur/mandible
RUN go get github.com/mattn/goveralls
RUN go get github.com/tools/godep
RUN go get golang.org/x/tools/cmd/vet
RUN godep restore
RUN godep go install -v .
CMD ["mandible"]
