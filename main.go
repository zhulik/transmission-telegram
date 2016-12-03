package main

import (
	"flag"
	"fmt"
	"github.com/pyed/transmission"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"os"
	"strings"
	"time"
)

const (
	VERSION = "2.0"

	HELP = `
	*list* or *li*
	Lists all the torrents, takes an optional argument which is a query to list only torrents that has a tracker matches the query, or some of it.

	*down* or *dl*
	Lists torrents with the status of Downloading or in the queue to download.

	*seeding* or *sd*
	Lists torrents with the status of Seeding or in the queue to seed.

	*paused* or *pa*
	Lists Paused torrents.

	*checking* or *ch*
	Lists torrents with the status of Verifying or in the queue to verify.

	*active* or *ac*
	Lists torrents that are actively uploading or downloading.

	*errors* or *er*
	Lists torrents with with errors along with the error message.

	*sort* or *so*
	Manipulate the sorting of the aforementioned commands, Call it without arguments for more.

	*trackers* or *tr*
	Lists all the trackers along with the number of torrents.

	*add* or *ad*
	Takes one or many URLs or magnets to add them, You can send a .torrent file via Telegram to add it.

	*search* or *se*
	Takes a query and lists torrents with matching names.

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

		// ignore anyone other than 'master'
		if strings.ToLower(update.Message.From.UserName) != strings.ToLower(masterUsername) {
			log.Printf("[INFO] Ignored a message from: %s", update.Message.From.String())
			continue
		}

		// tokenize the update
		tokens := strings.Split(update.Message.Text, " ")
		command := strings.ToLower(tokens[0])

		switch command {
		case "list", "/list", "li", "/li":
			go list(bot, client, update, tokens[1:])

		case "downs", "/downs", "dl", "/dl":
			go downs(bot, client, update)

		case "seeding", "/seeding", "sd", "/sd":
			go seeding(bot, client, update)

		case "paused", "/paused", "pa", "/pa":
			go paused(bot, client, update)

		case "checking", "/checking", "ch", "/ch":
			go checking(bot, client, update)

		case "active", "/active", "ac", "/ac":
			go active(bot, client, update)

		case "errors", "/errors", "er", "/er":
			go errors(bot, client, update)

		case "sort", "/sort", "so", "/so":
			go sort(bot, client, update, tokens[1:])

		case "trackers", "/trackers", "tr", "/tr":
			go trackers(bot, client, update)

		case "add", "/add", "ad", "/ad":
			go add(bot, client, update, tokens[1:])

		case "search", "/search", "se", "/se":
			go search(bot, client, update, tokens[1:])

		case "info", "/info", "in", "/in":
			go info(bot, client, update, tokens[1:])

		case "stop", "/stop", "sp", "/sp":
			go stop(bot, client, update, tokens[1:])

		case "start", "/start", "st", "/st":
			go start(bot, client, update, tokens[1:])

		case "check", "/check", "ck", "/ck":
			go check(bot, client, update, tokens[1:])

		case "stats", "/stats", "sa", "/sa":
			go stats(bot, client, update)

		case "speed", "/speed", "ss", "/ss":
			go speed(bot, client, update)

		case "count", "/count", "co", "/co":
			go count(bot, client, update)

		case "del", "/del":
			go del(bot, client, update, tokens[1:])

		case "deldata", "/deldata":
			go deldata(bot, client, update, tokens[1:])

		case "help", "/help":
			go send(bot, HELP, update.Message.Chat.ID, true)

		case "version", "/version":
			go version(bot, client, update)

		case "":
			// might be a file received
			go receiveTorrent(bot, client, botToken, update)

		default:
			// no such command, try help
			go send(bot, "no such command, try /help", update.Message.Chat.ID, false)
		}
	}
}
