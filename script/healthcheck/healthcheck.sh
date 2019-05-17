#!/bin/bash

me=$(basename $0)

msg() {
	echo 2>&1 $me: $@
}

die() {
	msg $@
	exit 1
}

app_home=/app
app=$app_home/gowebhello

cd $app_home || die "could not enter dir: $app_home"

# start service

restart() {
	msg restarting: $app
	pkill -9 gowebhello
	$app -quota=5 &
}

restart

# loop 

url=http://localhost:8080/www/

while :; do
	sleep 5
	http_code=$(curl -o /dev/null -s -I -X GET -w "%{http_code}" $url)
	exit_status=$?
	msg "exit_status=$exit_status http_code=$http_code"
	if [ "$exit_status" -ne 0 ] || [ "$http_code" != 200 ]; then
		restart
	fi
done


