FROM golang:alpine

RUN apk add make git

RUN mkdir -p /go/src/github.com/corneliusweig/ketall/

WORKDIR /go/src/github.com/corneliusweig/ketall/

CMD git clone --depth 1 https://github.com/corneliusweig/ketall.git . && \
    make all && \
    mv out/* /go/bin
