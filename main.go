package main

import (
	"flag"
	"fmt"
	"github.com/pyed/transmission"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

const (
	VERSION = "2.0"

	HELP = `
	*list* or *li*
	Lists all the torrents, takes an optional argument which is a query to list only torrents that has a tracker matches the query, or some of it.

	*downs* or *dl*
	Lists torrents with the status of Downloading or in the queue to download.

	*seeding* or *sd*
	Lists torrents with the status of Seeding or in the queue to seed.

	*paused* or *pa*
	Lists Paused torrents.

	*checking* or *ch*
	Lists torrents with the status of Verifying or in the queue to verify.

	*errors* or *er*
	Lists torrents with with errors along with the error message.

	*search* or *se*
	Takes a query and lists torrents with matching names.

	*sort* or *so*
	Manipulate the sorting of the aforementioned commands, Call it without arguments for more.

	*add* or *ad*
	Takes one or many URLs or magnets to add them, You can send a .torrent file via Telegram to add it.

	*info* or *in*
	Takes one or more torrent's IDs to list more info about them.

	*stop* or *sp*
	Takes one or more torrent's IDs to stop them, or _all_ to stop all torrents.

	*start* or *st*
	Takes one or more torrent's IDs to start them, or _all_ to start all torrents.

	*check* or *ck*
	Takes one or more torrent's IDs to verify them, or _all_ to verify all torrents.

	*del*
	Takes one or more torrent's IDs to delete them.

	*deldata*
	Takes one or more torrent's IDs to delete them and their data.

	*stats* or *sa*
	Shows Transmission's stats.

	*speed* or *ss*
	Shows the upload and download speeds.

	*count* or *co*
	Shows the torrents counts per status.

	*help*
	Shows this help message.

	*version*
	Shows version numbers.

	- Prefix commands with '/' if you want to talk to your bot in a group.
	- report any issues [here](https://github.com/pyed/transmission-telegram)
	`

	duration               = 60
	interval time.Duration = 2
)

func main() {
	var botToken string
	var masterUsername string
	var transmissionURL string
	var transmissionUsername string
	var transmissionPassword string
	var logFile string

	flag.StringVar(&botToken, "token", "", "Telegram bot token")
	flag.StringVar(&masterUsername, "master", "", "Your telegram handler, So the bot will only respond to you")
	flag.StringVar(&transmissionURL, "url", "http://localhost:9091/transmission/rpc", "Transmission RPC URL")
	flag.StringVar(&transmissionUsername, "username", "", "Transmission username")
	flag.StringVar(&transmissionPassword, "password", "", "Transmission password")
	flag.StringVar(&logFile, "logfile", "", "Send logs to a file")

	// set the usage message
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: transmission-bot -token=<TOKEN> -master=<@tuser> -url=[http://] -username=[user] -password=[pass]\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// make sure that we have the two madatory arguments: telegram token & master's handler.
	if botToken == "" ||
		masterUsername == "" {
		fmt.Fprintf(os.Stderr, "Error: Mandatory argument missing! (-token or -master)\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// make sure that the handler doesn't contain @
	masterUsername = strings.Replace(masterUsername, "@", "", -1)

	// if we got a log file, log to it
	if logFile != "" {
		logf, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(logf)
	}

	// if the `-username` flag isn't set, look into the environment variable 'TR_AUTH'
	if transmissionUsername == "" {
		if values := strings.Split(os.Getenv("TR_AUTH"), ":"); len(values) > 1 {
			transmissionUsername, transmissionPassword = values[0], values[1]
		}
	}

	// log the flags
	log.Printf("[INFO] Token=%s\nMaster=%s\nURL=%s\nUSER=%s\nPASS=%s",
		botToken, masterUsername, transmissionURL, transmissionUsername, transmissionPassword)

	client, err := transmission.New(transmissionURL, transmissionUsername, transmissionPassword)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Transmission: Make sure you have the right URL, Username and Password")
		os.Exit(1)
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Telegram: %s", err)
		os.Exit(1)
	}
	log.Printf("[INFO] Authorized: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Telegram: %s", err)
		os.Exit(1)
	}

	for update := range updates {
		// ignore edited messages
		if update.Message == nil {
			continue
		}

		// tokenize the update
		wrapper := WrapUpdate(update)

		// ignore anyone other than 'master'
		if strings.ToLower(update.Message.From.UserName) != strings.ToLower(masterUsername) {
			log.Printf("[INFO] Ignored a message from: %s", wrapper.Message.From.String())
			continue
		}

		go func() {
			defer func() {
				if recover() != nil {
					send(bot, "PANIC: something goes wrong...", wrapper.Message.Chat.ID)
					log.Println(string(debug.Stack()))
				}
			}()
			findHandler(wrapper.Command())(bot, client, wrapper)
		}()

	}
}

func findHandler(command string) CommandHandler {
	switch command {
	case "list", "/list", "li", "/li":
		return list

	case "downs", "/downs", "dl", "/dl":
		return downs

	case "seeding", "/seeding", "sd", "/sd":
		return seeding

	case "paused", "/paused", "pa", "/pa":
		return paused

	case "checking", "/checking", "ch", "/ch":
		return checking

	case "errors", "/errors", "er", "/er":
		return errors

	case "sort", "/sort", "so", "/so":
		return sort

	case "add", "/add", "ad", "/ad":
		return add

	case "search", "/search", "se", "/se":
		return search

	case "info", "/info", "in", "/in":
		return info

	case "stop", "/stop", "sp", "/sp":
		return stop

	case "start", "/start", "st", "/st":
		return start

	case "check", "/check", "ck", "/ck":
		return check

	case "stats", "/stats", "sa", "/sa":
		return stats

	case "speed", "/speed", "ss", "/ss":
		return speed

	case "count", "/count", "co", "/co":
		return count

	case "del", "/del":
		return del

	case "deldata", "/deldata":
		return deldata

	case "help", "/help":
		return help

	case "version", "/version":
		return version

	case "":
		return receiveTorrent

	default:
		return unknownCommand
	}
}
