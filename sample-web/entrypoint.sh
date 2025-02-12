#!/bin/sh
# Substitute variables in index.html
envsubst '$BASE_URL' < /usr/share/nginx/html/index.html.template > /usr/share/nginx/html/index.html
# Start nginx
nginx -g 'daemon off;'