#!/bin/sh

set -e

echo "Running certificate request script"

# Check that the certificate for $DOMAIN_NAME exists.
CERT_FILE=/etc/letsencrypt/live/$DOMAIN_NAME/fullchain.pem
if [ -e $CERT_FILE ]
then
  echo "Certificate located, not requesting"
else
  echo "Performing self-HTTP challenge before bothering LetsEncrypt"
  CHALLENGE_CHARS=$(openssl rand -hex 16)
  CHALLENGE_FNAME=$(openssl rand -hex 16).txt
  CHALLENGE_DIR=/var/www/acme/.well-known/acme-challenge
  CHALLENGE_FILE_PATH=$CHALLENGE_DIR/$CHALLENGE_FNAME

  mkdir -p $CHALLENGE_DIR
  echo $CHALLENGE_CHARS > $CHALLENGE_FILE_PATH
  RESP_CHARS=$(curl -sf http://$DOMAIN_NAME/.well-known/acme-challenge/$CHALLENGE_FNAME)
  if [ "$RESP_CHARS" = "$CHALLENGE_CHARS" ]
  then
    rm CHALLENGE_FILE_PATH
    echo "Internal challenge succeded"
    certbot certonly \
      --non-interactive \
      --agree-tos \
      --email $EMAIL \
      --keep-until-expiring \
      --webroot
      -w /var/www/acme

    echo "Successfully requested certificate for ${DOMAIN_NAME}"
    # Restart the nginx container, so that it picks up the presence of the certificate and loads it
    curl --unix-socket /var/run/docker.sock -X POST http:/v1.24/containers/$NGINX_CONTAINER_NAME/restart
  else
    echo "Internal challenge failed, please configure the domain ${DOMAIN_NAME} to point to this host"
    rm $CHALLENGE_FILE_PATH
  fi
fi