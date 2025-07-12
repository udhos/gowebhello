#!/bin/sh

gofmt -s -w ./gowebhello
go tool fix ./gowebhello
go vet ./gowebhello

hash gosimple && gosimple ./gowebhello
hash golint && golint ./gowebhello
hash staticcheck && staticcheck ./gowebhello

go test ./gowebhello
CGO_ENABLED=0 go install -ldflags='-s -w' -trimpath -v ./gowebhello
