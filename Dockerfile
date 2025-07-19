FROM golang:1.12
LABEL MAINTAINER="a nice guy"
WORKDIR $GOPATH/src/github.com/udhos/GOWEBHELLO
ADD . $GOPATH/src/github.com/udhos/GOWEBHELLO
RUN go build -o helloweb  gowebhello/main.go

ENTRYPOINT ["./helloweb"]