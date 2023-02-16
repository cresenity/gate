#!/bin/sh
envsubst '$${DEFAULT_TARGET}' <  /etc/nginx/templates/default.conf.template > /etc/nginx/conf.d/000-default.conf

nginx
./app/main