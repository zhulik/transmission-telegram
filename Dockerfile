FROM golang:1.8

RUN mkdir -p ./src/github.com/zhulik/transmission-telegram

WORKDIR ./src/github.com/zhulik/transmission-telegram

ADD . ./

RUN go get github.com/pyed/transmission

RUN go get github.com/dustin/go-humanize

RUN go get gopkg.in/telegram-bot-api.v4

RUN go get github.com/boltdb/bolt

RUN go build -o transmission-telegram ./


#ENTRYPOINT ["./transmission-telegram"]

CMD ["sh","-c","./transmission-telegram -token=$BOT_TOKEN -masters=$MASTERS -url=$TRANSMISSION_URL -username=$TRANSMISSION_USERNAME -password=$TRANSMISSION_PASSWORD"]