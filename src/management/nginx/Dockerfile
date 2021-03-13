FROM nginx:1.19.8-alpine

COPY ./start.sh /app/start.sh
COPY ./conf/* /etc/nginx/

RUN curl https://ssl-config.mozilla.org/ffdhe2048.txt > /etc/nginx/dhparam

EXPOSE 80 443

WORKDIR /app
ENTRYPOINT [ "./start.sh" ]