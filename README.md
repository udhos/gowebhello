[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/udhos/gowebhello/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/udhos/gowebhello)](https://goreportcard.com/report/github.com/udhos/gowebhello)

# gowebhello

gowebhello is a simple golang replacement for 'python -m SimpleHTTPServer'.
gowebhello can also be configured as an HTTPS web server using SSL certificates.

# Usage

## HTTPS

If you want to use TLS, you will need a certificate:

    $ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout key.pem -out cert.pem

## Building

### Without Modules, before Go 1.11

    # make sure GOPATH is either unset or properly set
    go get github.com/udhos/gowebhello
    go install github.com/udhos/gowebhello

### With Modules, starting from Go 1.11

    git clone https://github.com/udhos/gowebhello ;# clone outside of GOPATH
    cd gowebhello
    go install ./gowebhello

## Running

Use the '-h' switch to get command line help.

    $ gowebhello -h

## Example with TLS

Enable TLS by providing a certificate.
If you enable TLS, HTTP port will be redirected to HTTPS port.

    $ ~/go/bin/gowebhello
    2017/06/08 11:24:03 registering static directory /home/lab/go/src/github.com/udhos/gowebhello as www path /www/
    2017/06/08 11:24:03 using TCP ports HTTP=:8080 HTTPS=:8443 TLS=true
    2017/06/08 11:24:03 installing redirect from HTTP=:8080 to HTTPS=8443
    2017/06/08 11:24:03 serving HTTPS on TCP :8443

    Then open https://localhost:8443

## Example without TLS

If you do not provide a certificate, TLS will be disabled.

    $ ~/go/bin/gowebhello -cert=badcert
    2017/06/08 11:25:01 TLS cert file not found: badcert - disabling TLS
    2017/06/08 11:25:01 registering static directory /home/lab/go/src/github.com/udhos/gowebhello as www path /www/
    2017/06/08 11:25:01 using TCP ports HTTP=:8080 HTTPS=:8443 TLS=false
    2017/06/08 11:25:01 serving HTTP on TCP :8080

    Then open http://localhost:8080

## Example with HTTPS only

You can disable HTTP by specifying the same port to both -addr and -httpsAddr.

    $ ~/go/bin/gowebhello -addr :8443 -httpsAddr :8443
    2017/06/08 11:25:46 registering static directory /home/lab/go/src/github.com/udhos/gowebhello as www path /www/
    2017/06/08 11:25:46 using TCP ports HTTP=:8443 HTTPS=:8443 TLS=true
    2017/06/08 11:25:46 serving HTTPS on TCP :8443

# Container image

Find a small container image for gowebhello as `udhos/web-scratch` here:

https://hub.docker.com/r/udhos/web-scratch

#END
