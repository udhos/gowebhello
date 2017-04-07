#!/bin/sh

gofmt -s -w main.go
go tool fix main.go
go tool vet .
[ -x ~/go/bin/gosimple ] && ~/go/bin/gosimple main.go
[ -x ~/go/bin/golint ] && ~/go/bin/golint main.go
[ -x ~/go/bin/staticcheck ] && ~/go/bin/staticcheck main.go
go test github.com/udhos/gowebhello
go install github.com/udhos/gowebhello
