#!/bin/bash

home=/tmp/staticweb
log=$home/staticweb.log

mkdir $home

me=`basename $0`
msg() {
    echo >>$log $me: $*
}

redir() {
    while read i; do
        msg $i
    done
}

msg -- `date` begin

msg user: `id`
msg pwd: $PWD
msg home: $home

cd $home

wget -qO update-golang.sh  https://raw.githubusercontent.com/udhos/update-golang/master/update-golang.sh
chmod a+rx update-golang.sh

export DESTINATION=$PWD/golang
mkdir $DESTINATION
PROFILED=$HOME/.profile ./update-golang.sh


wget -qO main.go gowebhello.go https://raw.githubusercontent.com/udhos/gowebhello/master/main.go 

msg -- `date` end

nohup /usr/local/go/bin/go run main.go >>$log 2>&1 &


