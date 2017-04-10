FROM golang:1.8-alpine

RUN mkdir -p /go/src/app
WORKDIR /go/src/app
COPY . /go/src/app

RUN go-wrapper install

EXPOSE 8080 9080 9999

ENTRYPOINT ["app"]
