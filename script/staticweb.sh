#!/bin/bash

home=/home/ubuntu/static

[ -n "$1" ] && home=$1

mkdir $home

cd $home

wget -qO- https://raw.githubusercontent.com/udhos/update-golang/master/update-golang.sh | bash

wget -qO main.go gowebhello.go https://raw.githubusercontent.com/udhos/gowebhello/master/main.go 

nohup /usr/local/go/bin/go run main.go >log.txt 2>&1 &


