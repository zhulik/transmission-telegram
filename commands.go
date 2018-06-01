package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zhulik/transmission-telegram/settings"
	"gopkg.in/telegram-bot-api.v4"
	"os"
	"net/http"
	"io"
	"io/ioutil"
)

var (
	mainCommands = map[string]string{
		"stop all":  "StopAll",
		"stop":      "StopTorrent",
		"start all": "StartAll",
		"start":     "StartTorrent",
		"check all": "VerifyAll",
		"check":     "VerifyTorrent",
	}

	delParams = map[string]bool{
		"del":     false,
		"deldata": true,
	}
)

// receiveTorrent gets an update that potentially has a .torrent file to add
func receiveTorrent(bot telegramClient, client torrentClient, ud messageWrapper, s settings.Settings) {
	if ud.Document.FileID == "" {
		return // has no document
	}
	// get the file ID and make the config
	fconfig := tgbotapi.FileConfig{
		FileID: ud.Document.FileID,
	}
	file, err := bot.GetFile(fconfig)
	if err != nil {
		send(bot, fmt.Sprintf("*ERROR*: `%s`", err.Error()), ud.Chat.ID, true)
		return
	}
	// downloading file on Go side (not in Transmission) because of Transmission does not support proxy
	out, err := ioutil.TempFile("",file.FileID);
	if err != nil  {
		send(bot, fmt.Sprintf("*ERROR*: `%s`", err.Error()), ud.Chat.ID, true)
		return
	}
	defer out.Close()
	defer os.Remove(out.Name())

	// Get the data
	resp, err := http.Get(file.Link(bot.Token()))
	if err != nil {
		send(bot, fmt.Sprintf("*ERROR*: `%s`", err.Error()), ud.Chat.ID, true)
		return
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		send(bot, fmt.Sprintf("*ERROR*:bad status:  `%s`", resp.Status), ud.Chat.ID, true)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil  {
		send(bot, fmt.Sprintf("*ERROR*: `%s`", err.Error()), ud.Chat.ID, true)
		return
	}

	// add by local file
	addTorrentByFile(bot, client, ud,out.Name())
}

// stop takes id[s] of torrent[s] or 'all' to stop them
func mainCommand(bot telegramClient, client torrentClient, ud messageWrapper, s settings.Settings) {
	// make sure that we got at least one argument
	if len(ud.Tokens()) == 0 {
		send(bot, fmt.Sprintf("*%s*: needs an argument", ud.Command()), ud.Chat.ID, true)
		return
	}

	// if the first argument is 'all' then stop all torrents
	if ud.Tokens()[0] == "all" {
		if err := invokeError(client, mainCommands[fmt.Sprintf("%s all", ud.Command())]); err != nil {
			send(bot, fmt.Sprintf("*%s*: error occurred", ud.Command()), ud.Chat.ID, true)
			return
		}
		send(bot, fmt.Sprintf("*%s*: ok", ud.Command()), ud.Chat.ID, true)
		return
	}

	for _, id := range ud.Tokens() {
		num, err := strconv.Atoi(id)
		if err != nil {
			send(bot, fmt.Sprintf("*%s*: `%s` is not a number", ud.Command(), id), ud.Chat.ID, true)
			continue
		}
		status, err := invokeStatus(client, mainCommands[ud.Command()], num)
		if err != nil {
			send(bot, fmt.Sprintf("*%s*: `%s`", ud.Command(), err.Error()), ud.Chat.ID, true)
			continue
		}

		torrent, err := client.GetTorrent(num)
		if err != nil {
			send(bot, fmt.Sprintf("*[fail] %s*: No torrent with an ID of %d", ud.Command(), num), ud.Chat.ID, true)
			return
		}
		send(bot, fmt.Sprintf("*[%s] %s*: `%s`", status, ud.Command(), torrent.Name), ud.Chat.ID, true)
	}
}

// del takes an id or more, and delete the corresponding torrent/s
func delCommand(bot telegramClient, client torrentClient, ud messageWrapper, s settings.Settings) {
	// make sure that we got an argument
	if len(ud.Tokens()) == 0 {
		send(bot, fmt.Sprintf("*%s*: needs an ID", ud.Command()), ud.Chat.ID, true)
		return
	}

	// loop over ud.Tokens() to read each potential id
	for _, id := range ud.Tokens() {
		num, err := strconv.Atoi(id)
		if err != nil {
			send(bot, fmt.Sprintf("*%s*: `%s` is not an ID", ud.Command(), id), ud.Chat.ID, true)
			return
		}

		name, err := client.DeleteTorrent(num, delParams[ud.Command()])
		if err != nil {
			send(bot, fmt.Sprintf("*%s*: `%s`", ud.Command(), err.Error()), ud.Chat.ID, true)
			return
		}

		send(bot, fmt.Sprintf("*%s*: `%s`", ud.Command(), name), ud.Chat.ID, true)
	}
}

// version sends transmission version + transmission-telegram version
func version(bot telegramClient, client torrentClient, ud messageWrapper, s settings.Settings) {
	send(bot, fmt.Sprintf("Transmission *%s*\nTransmission-telegram *%s*", client.Version(), VERSION), ud.Chat.ID, true)
}

// addTorrentsByURL adds torrent files or magnet links passed by rls
func addTorrentsByURL(bot telegramClient, client torrentClient, ud messageWrapper, urls []string) {
	if len(urls) == 0 {
		send(bot, "*add*: needs atleast one URL", ud.Chat.ID, true)
		return
	}

	// loop over the URL/s and add them
	for _, url := range urls {
		torrent, err := client.AddByURL(url)
		if err != nil {
			send(bot, fmt.Sprintf("*add*: `%s`", err.Error()), ud.Chat.ID, true)
			continue
		}

		// check if torrent.Name is empty, then an error happened
		if torrent.Name == "" {
			send(bot, fmt.Sprintf("*add*: error adding `%s`", url), ud.Chat.ID, true)
			continue
		}
		send(bot, fmt.Sprintf("*add*: *%d* `%s`", torrent.ID, torrent.Name), ud.Chat.ID, true)
	}
}

func addTorrentByFile(bot telegramClient, client torrentClient, ud messageWrapper, fileName string){
	torrent, err := client.AddByLocalFile(fileName)
	if err != nil {
		send(bot, fmt.Sprintf("*add*: `%s`", err.Error()), ud.Chat.ID, true)
		return
	}

	// check if torrent.Name is empty, then an error happened
	if torrent.Name == "" {
		send(bot, fmt.Sprintf("*add*: error adding `%s`", fileName), ud.Chat.ID, true)
		return
	}
	send(bot, fmt.Sprintf("*add*: *%d* `%s`", torrent.ID, torrent.Name), ud.Chat.ID, true)
}

// add takes an URL to a .torrent file in message to add it to transmission
func add(bot telegramClient, client torrentClient, ud messageWrapper, s settings.Settings) {
	addTorrentsByURL(bot, client, ud, ud.Tokens())
}

// help sends help messsage
func help(bot telegramClient, client torrentClient, ud messageWrapper, s settings.Settings) {
	send(bot, HELP, ud.Chat.ID, true)
}

// unknownCommand sends message that command is unknown
func unknownCommand(bot telegramClient, client torrentClient, ud messageWrapper, s settings.Settings) {
	send(bot, "no such command, try /help", ud.Chat.ID, true)
}

// sort changes torrents sorting
func sortCommand(bot telegramClient, client torrentClient, ud messageWrapper, s settings.Settings) {
	if len(ud.Tokens()) == 0 {
		send(bot, `sort takes one of:
			(*id, name, age, size, progress, downspeed, upspeed, download, upload, ratio*)
			optionally start with (*rev*) for reversed order
			e.g. "*sort rev size*" to get biggest torrents first.`, ud.Chat.ID, true)
		return
	}

	var reversed bool
	tokens := ud.Tokens()
	if strings.ToLower(ud.Tokens()[0]) == "rev" {
		reversed = true
		tokens = ud.Tokens()[1:]
	}

	mode := sortMethod{strings.ToLower(tokens[0]), reversed}

	if mode, ok := sortingMethods[mode]; ok {
		client.SetSort(mode)
		send(bot, fmt.Sprintf("*sort*: `%s` reversed: %t", tokens[0], reversed), ud.Chat.ID, true)
	} else {
		send(bot, "*sort*: unkown sorting method", ud.Chat.ID, true)
	}
}
