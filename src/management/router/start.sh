#!/bin/sh

set -e

# Write the certificate configuration for the provided domain, but don't consume it unless
# the certificate is present on disk.
DOMAIN_DIR=/etc/letsencrypt/live/$DOMAIN_NAME
printf "ssl_certificate ${DOMAIN_DIR}/fullchain.pem;\nssl_certificate_key ${DOMAIN_DIR}/privkey.pem;\n" > /etc/nginx/certs.conf

# If the certificate exists, overwrite the empty ssl.conf file with the completed version
if [ -e $DOMAIN_DIR/fullchain.pem ]
then
  cp /etc/nginx/ssl-template.conf /etc/nginx/ssl.conf
fi

exec nginx -g "daemon off;" -c /etc/nginx/nginx.conf