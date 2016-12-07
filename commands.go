package main

import (
	"fmt"
	"github.com/pyed/transmission"
	"gopkg.in/telegram-bot-api.v4"
	"strconv"
	"strings"
)

var (
	MainCommands = map[string]string{
		"stop all":  "StopAll",
		"stop":      "StopTorrent",
		"start all": "StartAll",
		"start":     "StartTorrent",
		"check all": "VerifyAll",
		"check":     "VerifyTorrent",
	}

	DelParams = map[string]bool{
		"del":     false,
		"deldata": true,
	}
)

// receiveTorrent gets an update that potentially has a .torrent file to add
func receiveTorrent(bot *tgbotapi.BotAPI, client TransmissionClient, ud MessageWrapper) {
	if ud.Document.FileID == "" {
		return // has no document
	}

	// get the file ID and make the config
	fconfig := tgbotapi.FileConfig{
		FileID: ud.Document.FileID,
	}
	file, err := bot.GetFile(fconfig)
	if err != nil {
		send(bot, fmt.Sprintf("*ERROR*: `%s`", err.Error()), ud.Chat.ID)
		return
	}

	// add by file URL
	addTorrentsByURL(bot, client, ud, []string{file.Link(bot.Token)})
}

// stop takes id[s] of torrent[s] or 'all' to stop them
func mainCommand(bot *tgbotapi.BotAPI, client TransmissionClient, ud MessageWrapper) {
	// make sure that we got at least one argument
	if len(ud.Tokens()) == 0 {
		send(bot, fmt.Sprintf("*%s*: needs an argument", ud.Command()), ud.Chat.ID)
		return
	}

	// if the first argument is 'all' then stop all torrents
	if ud.Tokens()[0] == "all" {
		if err := InvokeError(client, MainCommands[fmt.Sprintf("%s all", ud.Command())]); err != nil {
			send(bot, fmt.Sprintf("*%s*: error occurred", ud.Command()), ud.Chat.ID)
			return
		}
		send(bot, fmt.Sprintf("*%s*: ok", ud.Command()), ud.Chat.ID)
		return
	}

	for _, id := range ud.Tokens() {
		num, err := strconv.Atoi(id)
		if err != nil {
			send(bot, fmt.Sprintf("*%s*: `%s` is not a number", ud.Command(), id), ud.Chat.ID)
			continue
		}
		status, err := InvokeStatus(client, MainCommands[ud.Command()], num)
		if err != nil {
			send(bot, fmt.Sprintf("*%s*: `%s`", ud.Command(), err.Error()), ud.Chat.ID)
			continue
		}

		torrent, err := client.GetTorrent(num)
		if err != nil {
			send(bot, fmt.Sprintf("*[fail] %s*: No torrent with an ID of %d", ud.Command(), num), ud.Chat.ID)
			return
		}
		send(bot, fmt.Sprintf("*[%s] %s*: `%s`", status, ud.Command(), torrent.Name), ud.Chat.ID)
	}
}

// del takes an id or more, and delete the corresponding torrent/s
func delCommand(bot *tgbotapi.BotAPI, client TransmissionClient, ud MessageWrapper) {
	// make sure that we got an argument
	if len(ud.Tokens()) == 0 {
		send(bot, fmt.Sprintf("*%s*: needs an ID", ud.Command()), ud.Chat.ID)
		return
	}

	// loop over ud.Tokens() to read each potential id
	for _, id := range ud.Tokens() {
		num, err := strconv.Atoi(id)
		if err != nil {
			send(bot, fmt.Sprintf("*%s*: `%s` is not an ID", ud.Command(), id), ud.Chat.ID)
			return
		}

		name, err := client.DeleteTorrent(num, DelParams[ud.Command()])
		if err != nil {
			send(bot, fmt.Sprintf("*%s*: `%s`", ud.Command(), err.Error()), ud.Chat.ID)
			return
		}

		send(bot, fmt.Sprintf("*%s*: `%s`", ud.Command(), name), ud.Chat.ID)
	}
}

// version sends transmission version + transmission-telegram version
func version(bot *tgbotapi.BotAPI, client TransmissionClient, ud MessageWrapper) {
	send(bot, fmt.Sprintf("Transmission *%s*\nTransmission-telegram *%s*", client.Version(), VERSION), ud.Chat.ID)
}

// addTorrentsByURL adds torrent files or magnet links passed by rls
func addTorrentsByURL(bot *tgbotapi.BotAPI, client TransmissionClient, ud MessageWrapper, urls []string) {
	if len(urls) == 0 {
		send(bot, "*add*: needs atleast one URL", ud.Chat.ID)
		return
	}

	// loop over the URL/s and add them
	for _, url := range urls {
		cmd := transmission.NewAddCmdByURL(url)

		torrent, err := client.ExecuteAddCommand(cmd)
		if err != nil {
			send(bot, fmt.Sprintf("*add*: `%s`", err.Error()), ud.Chat.ID)
			continue
		}

		// check if torrent.Name is empty, then an error happened
		if torrent.Name == "" {
			send(bot, fmt.Sprintf("*add*: error adding `%s`", url), ud.Chat.ID)
			continue
		}
		send(bot, fmt.Sprintf("*add*: *%d* `%s`", torrent.ID, torrent.Name), ud.Chat.ID)
	}
}

// add takes an URL to a .torrent file in message to add it to transmission
func add(bot *tgbotapi.BotAPI, client TransmissionClient, ud MessageWrapper) {
	addTorrentsByURL(bot, client, ud, ud.Tokens())
}

// help sends help messsage
func help(bot *tgbotapi.BotAPI, client TransmissionClient, ud MessageWrapper) {
	send(bot, HELP, ud.Chat.ID)
}

// unknownCommand sends message that command is unknown
func unknownCommand(bot *tgbotapi.BotAPI, client TransmissionClient, ud MessageWrapper) {
	send(bot, "no such command, try /help", ud.Chat.ID)
}

// sort changes torrents sorting
func sort(bot *tgbotapi.BotAPI, client TransmissionClient, ud MessageWrapper) {
	if len(ud.Tokens()) == 0 {
		send(bot, `sort takes one of:
			(*id, name, age, size, progress, downspeed, upspeed, download, upload, ratio*)
			optionally start with (*rev*) for reversed order
			e.g. "*sort rev size*" to get biggest torrents first.`, ud.Chat.ID)
		return
	}

	var reversed bool
	tokens := ud.Tokens()
	if strings.ToLower(ud.Tokens()[0]) == "rev" {
		reversed = true
		tokens = ud.Tokens()[1:]
	}

	mode := SortMethod{strings.ToLower(tokens[0]), reversed}

	if mode, ok := SortingMethods[mode]; ok {
		client.SetSort(mode)
		send(bot, fmt.Sprintf("*sort*: `%s` reversed: %t", tokens[0], reversed), ud.Chat.ID)
	} else {
		send(bot, "*sort*: unkown sorting method", ud.Chat.ID)
	}
}
