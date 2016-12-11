package main

import (
	"flag"
	"fmt"
	"github.com/pyed/transmission"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"os"
	"runtime/debug"
	"sort"
	"strings"
)

const (
	VERSION = "2.0"

	HELP = `
	*ls* [dl, sd, pa, ch, er]
	Lists the torrents. Optional argument:
		*dl* - Lists torrents with the status of Downloading or in the queue to download.
		*sd* - Lists torrents with the status of Seeding or in the queue to seed.
		*pa* - Lists Paused torrents.
		*ch* - Lists torrents with the status of Verifying or in the queue to verify.
		*er* - Lists torrents with with errors along with the error message.

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
)

func main() {
	var botToken string
	var masterUsernames string
	var transmissionURL string
	var transmissionUsername string
	var transmissionPassword string
	var logFile string

	flag.StringVar(&botToken, "token", "", "Telegram bot token")
	flag.StringVar(&masterUsernames, "masters", "", "Your telegram handlers, separated with comma, so the bot will only respond to them")
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
	if botToken == "" || masterUsernames == "" {
		fmt.Fprintf(os.Stderr, "Error: Mandatory argument missing! (-token or -masters)\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// make sure that the handler doesn't contain @
	masterUsernames = strings.ToLower(strings.Replace(masterUsernames, "@", "", -1))

	masters := strings.Split(masterUsernames, ",")
	sort.Strings(masters)

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
	log.Printf("[INFO] Token=%s\nMasters=%s\nURL=%s\nUSER=%s\nPASS=%s",
		botToken, masterUsernames, transmissionURL, transmissionUsername, transmissionPassword)

	client, err := transmission.New(transmissionURL, transmissionUsername, transmissionPassword)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Transmission: Make sure you have the right URL, Username and Password")
		os.Exit(1)
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	// bot.Debug = true
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Telegram: %s", err)
		os.Exit(1)
	}
	log.Printf("[INFO] Authorized: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Printf("[ERROR] Telegram: %s", err)
		os.Exit(1)
	}

	b := &TelegramClientWrapper{bot: bot}
	// go notifyFinished(b, client, masters)

	for update := range updates {
		var wrapper MessageWrapper
		if update.Message == nil {
			if update.UpdateID > 0 {
				log.Println(update.EditedMessage.Text)
				wrapper = WrapMessage(update.EditedMessage)
			} else {
				continue
			}
		} else {
			wrapper = WrapMessage(update.Message)
		}

		// ignore anyone other than 'masters'
		if sort.SearchStrings(masters, strings.ToLower(wrapper.From.UserName)) == len(masters) {
			log.Printf("[INFO] Ignored a message from: %s", wrapper.Message.From.String())
			continue
		}

		go func() {
			defer func() {
				if recover() != nil {
					send(b, "PANIC: something goes wrong...", wrapper.Message.Chat.ID)
					log.Println(string(debug.Stack()))
				}
			}()
			findHandler(wrapper.Command())(b, client, wrapper)
		}()

	}
}

func findHandler(command string) CommandHandler {
	switch command {
	case "ls", "/ls":
		return list

	case "sort", "/sort", "so", "/so":
		return sortComand

	case "add", "/add", "ad", "/ad":
		return add

	case "search", "/search", "se", "/se":
		return search

	case "info", "/info", "in", "/in":
		return info

	case "stop", "/stop", "sp", "/sp", "start", "/start", "st", "/st", "check", "/check", "ck", "/ck":
		return mainCommand

	case "stats", "/stats", "sa", "/sa":
		return stats

	case "progress", "/progress", "pr", "/pr":
		return progress

	case "speed", "/speed", "ss", "/ss":
		return speed

	case "count", "/count", "co", "/co":
		return count

	case "del", "/del", "deldata", "/deldata":
		return delCommand

	case "notify":
		return notifyFinished

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
