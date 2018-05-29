# transmission-telegram

Manage your transmission through Telegram. Easyly add downloads with magnet links, control your
transmission even if your server doesn't have white IP address.

## Install

`go get -u github.com/zhulik/transmission-telegram`

## Usage

`transmission-telegram -masters <user,other_user> -token <your token> -username <transmission username> -password <transmission password> -url http://<host:port>/transmission/rpc`
## Docker usage

`docker build --tag=transmission-telegram .`

`docker run -e BOT_TOKEN=<token> -e MASTERS=<MyUser> -e TRANSMISSION_URL=http://host:9091/transmission/rpc transmission-telegram`

Take a look at [docker-compose.yml](docker-compose.yml) for full reference of supported environment variables
## Available commands and queries
All available queries you can view with `help` command. There are 3 types of commands:

* Simple queries - one query, one response(ex. list, search)
* Continuous queries - queries wich updates previously sent message(ex. info, progress)
* Commands - actions that can change daemon state(ex. add, del)



## Todo

* Manage files and locations
* Keyboards and buttons
* Persistent settings storage
