#!/bin/sh

gofmt -s -w ./gowebhello
go tool fix ./gowebhello
go tool vet ./gowebhello

hash gosimple && gosimple ./gowebhello
hash golint && golint ./gowebhello
hash staticcheck && staticcheck ./gowebhello

go test ./gowebhello
go install -v ./gowebhello
