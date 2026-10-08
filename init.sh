#!/bin/sh
export RESOLVER_ADDRESS="${RESOLVER_ADDRESS:-127.0.0.11}"

envsubst '$${RESOLVER_ADDRESS}' < /etc/nginx/templates/nginx.conf.template > /etc/nginx/nginx.conf
envsubst '$${DEFAULT_TARGET}' <  /etc/nginx/templates/default.conf.template > /etc/nginx/conf.d/000-default.conf

nginx
./app/main