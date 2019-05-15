#!/bin/bash

me=$(basename $0)

msg() {
	echo >&2 $me: $@
}

die() {
	msg $@
	exit 1
}

do_install() {
	local hc=healthcheck/healthcheck.sh
	[ -f $hc ] || die missing executable: $hc
	[ -r $hc ] || die missing executable: $hc
	[ -x $hc ] || die missing executable: $hc

	[ -f /app/healthcheck.sh ] && die already installed, use: $0 remove

	cp $hc                             /app
	cp healthcheck/healthcheck.service /lib/systemd/system

	systemctl daemon-reload
	systemctl enable healthcheck.service
	systemctl reload-or-restart healthcheck.service

	msg check service: systemctl status healthcheck
	msg check logs:    journalctl -u healthcheck -f
}

do_uninstall() {
	systemctl daemon-reload
	systemctl stop healthcheck.service
	systemctl disable healthcheck.service
	pkill -9 healthcheck.sh
	rm /app/healthcheck.sh
}

case "$1" in
	remove)
		do_uninstall
	;;
	'')
		do_install
	;;
	*)
		msg invalid argument: "$1"
		echo >&2 usage: $0 [remove]
		exit 2
	;;
esac

