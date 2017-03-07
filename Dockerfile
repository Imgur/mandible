FROM golang:1.8-stretch
RUN apt-get update && apt-get install -yqq aspell aspell-en libaspell-dev tesseract-ocr tesseract-ocr-eng imagemagick optipng exiftool libjpeg-progs webp
ADD docker/meme.traineddata /usr/share/tesseract-ocr/tessdata/meme.traineddata
ADD docker/imagemagick_policy.xml /etc/ImageMagick-6/policy.xml
RUN mkdir -p /etc/mandible /tmp/imagestore
ENV MANDIBLE_CONF /etc/mandible/conf.json
ENV GO15VENDOREXPERIMENT 1
ADD . /go/src/github.com/Imgur/mandible
WORKDIR /go/src/github.com/Imgur/mandible

RUN wget "https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-1.0.0.tar.gz"
RUN tar -C /usr/local -xvf libtensorflow-cpu-linux-x86_64-1.0.0.tar.gz
ENV LIBRARY_PATH=/usr/local/lib
ENV LD_LIBRARY_PATH=/usr/local/lib

RUN go get github.com/mattn/goveralls
RUN go get github.com/tools/godep
RUN godep restore
RUN godep go install -v .
CMD ["mandible"]
