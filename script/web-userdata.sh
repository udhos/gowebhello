#!/bin/bash

[ -d /web ] || mkdir /web
cd /web

if [ ! -f gowebhello ]; then
	wget -O gowebhello https://github.com/udhos/gowebhello/releases/download/v0.5/gowebhello_linux_amd64
fi

chmod a+rx gowebhello

cat >/lib/systemd/system/web.service <<__EOF__
[Unit]
Description=Gowebhello Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/web
ExecStart=/web/gowebhello
Restart=on-failure

[Install]
WantedBy=multi-user.target
__EOF__

systemctl daemon-reload
systemctl enable web.service
systemctl reload-or-restart web.service
