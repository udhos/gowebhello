#!/bin/bash

die() {
    echo >&2 $0: $*
    exit 1
}

if [ -z "$home" ]; then
    home=/gowebhello
fi

main=$home/main.go

echo $0: $main

[ -r "$main" ] || die "unable to read: $main"

# kill previous instance
pkill -f -9 '^.+go-build.+main$'

# start new instance
nohup /usr/local/go/bin/go run $main &

