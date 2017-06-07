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
    
Example with TLS:

    $ ~/go/bin/gowebhello 
    2017/06/07 01:20:06 registering static directory /home/everton/go/src/github.com/udhos/gowebhello as www path /www/
    2017/06/07 01:20:06 serving on port TCP HTTP=:8080 HTTPS=:8443 TLS=true
    2017/06/07 01:20:06 installing redirect from HTTP=:8080 to HTTPS=8443

    Then open https://localhost:8443

Example without TLS:

    $ ~/go/bin/gowebhello 
    2017/06/07 01:24:45 TLS key file not found: key.pem - disabling TLS
    2017/06/07 01:24:45 TLS cert file not found: cert.pem - disabling TLS
    2017/06/07 01:24:45 registering static directory /home/everton/go/src/github.com/udhos/gowebhello as www path /www/
    2017/06/07 01:24:45 serving on port TCP HTTP=:8080 HTTPS=:8443 TLS=false

    Then open http://localhost:8080

END
===
