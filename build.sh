#!/bin/bash

go install golang.org/x/vuln/cmd/govulncheck@latest
go install golang.org/x/tools/cmd/deadcode@latest
go install github.com/mgechev/revive@latest

gofmt -s -w .

revive ./...

echo gocyclo begin
gocyclo -over 15 .
echo gocyclo end

go mod tidy

govulncheck ./...

deadcode ./gowebhello

go env -w CGO_ENABLED=1

go test -race ./...

go env -w CGO_ENABLED=0

go install -ldflags='-s -w' -trimpath -v ./gowebhello

go env -u CGO_ENABLED
