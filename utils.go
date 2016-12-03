package main

import (
	"github.com/pyed/transmission"
	"gopkg.in/telegram-bot-api.v4"
	"strings"
	"unicode/utf8"
)

type UpdateWrapper struct {
	tgbotapi.Update
	command string
	tokens  []string
}

func WrapUpdate(update tgbotapi.Update) UpdateWrapper {
	tokens := strings.Split(update.Message.Text, " ")
	command := tokens[0]
	args := tokens[1:]
	return UpdateWrapper{update, command, args}
}

func (w UpdateWrapper) Command() string {
	return w.command
}

func (w UpdateWrapper) Tokens() []string {
	return w.tokens
}

type CommandHandler func(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper)
type TorrentFilter func(torrent *transmission.Torrent) bool

func ellipsisString(str string, length int) string {
	if utf8.RuneCountInString(str) > length {
		return string([]rune(str)[:length-3]) + "..."
	}
	return str
}

func sendFilteredTorrets(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper, filter TorrentFilter) {
	torrents, err := client.GetTorrents()
	if err != nil {
		send(bot, "Torrents obtain error: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	filteredTorrents := transmission.Torrents{}
	for _, torrent := range torrents {
		if filter(torrent) {
			filteredTorrents = append(filteredTorrents, torrent)
		}
	}
	sendTorrents(bot, ud, filteredTorrents)
}
