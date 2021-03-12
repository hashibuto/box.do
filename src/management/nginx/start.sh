#!/bin/sh

set -e

exec nginx -g "daemon off;" -c /etc/nginx/nginx.conf