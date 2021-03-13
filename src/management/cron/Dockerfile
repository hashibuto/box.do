FROM alpine:3.13.2

RUN apk update && \
  apk add --no-cache \
  certbot \
  curl \
  openssl

COPY /src/start.sh /app/start.sh
COPY /src/scripts/* /app/scripts/
COPY /src/crontabs/* /etc/crontabs/

WORKDIR /app

CMD [ "./start.sh" ]