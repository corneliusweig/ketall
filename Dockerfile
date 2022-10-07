FROM golang:alpine

RUN apk add make git

RUN mkdir -p /go/src/github.com/flanksource/ketall/

WORKDIR /go/src/github.com/flanksource/ketall/

CMD git clone --depth 1 https://github.com/flanksource/ketall.git . && \
    make all && \
    mv out/* /go/bin
