FROM golang:alpine as build-stage

WORKDIR /app
COPY . .
RUN go build -o main


FROM nginx:alpine as production-stage

COPY ./nginx/default.conf.template /etc/nginx/templates/
COPY ./nginx/nginx.conf.template /etc/nginx/nginx.conf
COPY --from=build-stage /app/main /app/
COPY init.sh /usr/local/bin

RUN chmod +x /usr/local/bin/init.sh
RUN apk add certbot-nginx
RUN echo "0 0 1 * * /usr/bin/certbot renew --quiet" >> /etc/crontabs/root

ENTRYPOINT [ "/usr/local/bin/init.sh" ]
