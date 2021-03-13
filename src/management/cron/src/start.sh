#!/bin/sh

set -e

echo "Running startup job(s)"

/app/scripts/request-cert.sh
chmod 0600 /etc/crontabs/*

echo "Starting CRON daemon"
crond -f -d 8
