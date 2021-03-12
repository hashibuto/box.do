FROM alpine:3.13.2

RUN apk update && \
  apk add --no-cache \
  certbot

COPY /src/startup.sh /app/startup.sh

WORKDIR /app

CMD [ "./startup.sh" ]