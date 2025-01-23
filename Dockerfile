FROM golang:alpine

RUN apk add make git

RUN mkdir -p /go/src/github.com/stackitcloud/kubectl-get-all/

WORKDIR /go/src/github.com/stackitcloud/kubectl-get-all/

CMD git clone --depth 1 https://github.com/stackitcloud/kubectl-get-all.git . && \
    make all && \
    mv out/* /go/bin
