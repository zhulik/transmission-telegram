FROM golang:1.13.4-alpine AS builder
ENV APPDIR /app
RUN mkdir $APPDIR
WORKDIR $APPDIR
RUN apk add --no-cache make
ADD go.mod go.sum Makefile $APPDIR/
COPY . $APPDIR
RUN make

FROM alpine:latest
ENV APPDIR /app
RUN mkdir $APPDIR
WORKDIR $APPDIR
COPY --from=builder $APPDIR/transmission-telegram .
CMD ["sh","-c","./transmission-telegram -token=$BOT_TOKEN -masters=$MASTERS -url=$TRANSMISSION_URL -username=$TRANSMISSION_USERNAME -password=$TRANSMISSION_PASSWORD"]
