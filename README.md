[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/udhos/gowebhello/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/udhos/gowebhello)](https://goreportcard.com/report/github.com/udhos/gowebhello)

# gowebhello
gowebhello is a simple golang replacement for 'python -m SimpleHTTPServer'.

Usage
=====

If you want to use TLS, you will need a certificate:

    $ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout key.pem -out cert.pem

Build, install and run:

    $ export GOPATH=~/go ;# not needed since go1.8
    $ go get github.com/udhos/gowebhello
    $ go install github.com/udhos/gowebhello
    $ ~/go/bin/gowebhello
    2017/04/07 18:20:14 registering static directory /home/lab/go/src/github.com/udhos/gowebhello as www path /www/
    2017/04/07 18:20:14 serving on port TCP :8080

Then open http://localhost:8080, or https://localhost:8080 for TLS.

END
===
