#!/bin/sh

set -e

echo "Running startup job(s)"

echo "Starting CRON daemon"
crond -f -d 8
