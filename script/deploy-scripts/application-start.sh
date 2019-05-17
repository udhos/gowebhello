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
pkill -9 -f '^.+go-build.+main$'

# start new instance
#nohup /usr/local/go/bin/go run $main &
# http://docs.aws.amazon.com/codedeploy/latest/userguide/troubleshooting-deployments.html#troubleshooting-long-running-processes
/usr/local/go/bin/go run $main >/dev/null 2>/dev/null </dev/null &

