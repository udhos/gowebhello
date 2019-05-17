#!/bin/bash

#APP_URL=https://github.com/udhos/gowebhello/releases/download/v0.6/gowebhello_linux_amd64

echo >&2 "env var APP_URL=[$APP_URL]"

if [ -z "$APP_URL" ]; then
	echo >&2 "missing env var APP_URL=[$APP_URL]"
	exit 1
fi

app_dir=/web

[ -d $app_dir ] || mkdir $app_dir
cd $app_dir || echo >&2 "could not cd: app_dir=$app_dir"

[ -f gowebhello ] || curl -o gowebhello "$APP_URL"

chmod a+rx gowebhello

#
# web service
#

cat >/lib/systemd/system/web.service <<__EOF__
[Unit]
Description=Gowebhello Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$app_dir
ExecStart=$app_dir/gowebhello -burnCpu -quotaTime 1m
Restart=on-failure

[Install]
WantedBy=multi-user.target
__EOF__

#
# healthcheck script
#

cat >$app_dir/healthcheck.sh <<'__EOF__' ;# caution: quotes force full here-doc as literal
#!/bin/bash

url=http://localhost:8080/www/

while :; do
        sleep 5
        http_code=$(curl -o /dev/null -s -I -X GET -w '%{http_code}' "$url")
        exit_status=$?
        echo >&2 "exit_status=$exit_status http_code=$http_code"
        if [ "$exit_status" -ne 0 ] || [ "$http_code" != 200 ]; then
        	echo >&2 restarting: systemctl restart web.service
		systemctl restart web.service
        fi
done
__EOF__

chmod a+rx $app_dir/healthcheck.sh

#
# healthcheck service
#

cat >/lib/systemd/system/healthcheck.service <<__EOF__
#!/bin/bash

[Unit]
Description=Health Check Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$app_dir
ExecStart=$app_dir/healthcheck.sh
Restart=on-failure

[Install]
WantedBy=multi-user.target
__EOF__

systemctl daemon-reload
systemctl enable web.service
systemctl restart web.service
systemctl enable healthcheck.service
systemctl restart healthcheck.service

echo "check service: systemctl status web healthcheck"
echo "check logs:    journalctl -u web -u healthcheck -f"
