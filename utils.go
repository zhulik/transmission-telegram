package main

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/dustin/go-humanize"
	"github.com/pyed/transmission"
	"github.com/zhulik/transmission-telegram/settings"
	"gopkg.in/telegram-bot-api.v4"
)

var (
	sortingMethods = map[sortMethod]transmission.Sorting{
		sortMethod{"id", false}:       transmission.SortID,
		sortMethod{"id", true}:        transmission.SortRevID,
		sortMethod{"name", false}:     transmission.SortName,
		sortMethod{"name", true}:      transmission.SortRevName,
		sortMethod{"age", false}:      transmission.SortAge,
		sortMethod{"age", true}:       transmission.SortRevAge,
		sortMethod{"size", false}:     transmission.SortSize,
		sortMethod{"size", true}:      transmission.SortRevSize,
		sortMethod{"progress", false}: transmission.SortProgress,
		sortMethod{"progress", true}:  transmission.SortRevProgress,
		sortMethod{"downsped", false}: transmission.SortDownSpeed,
		sortMethod{"downsped", true}:  transmission.SortRevDownSpeed,
		sortMethod{"upspeed", false}:  transmission.SortUpSpeed,
		sortMethod{"upspeed", true}:   transmission.SortRevUpSpeed,

		sortMethod{"download", false}: transmission.SortDownloaded,
		sortMethod{"download", true}:  transmission.SortRevDownloaded,

		sortMethod{"upload", false}: transmission.SortUploaded,
		sortMethod{"upload", true}:  transmission.SortRevUploaded,

		sortMethod{"ratio", false}: transmission.SortRatio,
		sortMethod{"ratio", true}:  transmission.SortRevRatio,
	}
)

type sortMethod struct {
	name     string
	reversed bool
}

type messageWrapper struct {
	*tgbotapi.Message
	command string
	tokens  []string
}

func wrapMessage(message *tgbotapi.Message) messageWrapper {
	tokens := strings.Split(message.Text, " ")
	command := strings.ToLower(tokens[0])
	args := tokens[1:]
	return messageWrapper{message, command, args}
}

func (w messageWrapper) Command() string {
	return w.command
}

func (w messageWrapper) Tokens() []string {
	return w.tokens
}

type commandHandler func(bot telegramClient, client transmissionClient, ud messageWrapper, s settings.Settings)
type torrentFilter func(torrent *transmission.Torrent) bool

func ellipsisString(str string, length int) string {
	if utf8.RuneCountInString(str) > length {
		return string([]rune(str)[:length-3]) + "..."
	}
	return str
}

// send takes a chat id and a message to send, returns the message id of the send message
func send(bot telegramClient, text string, chatID int64, addKeyboard bool) int {
	var keyboard interface{}

	if addKeyboard {
		keyboard = commandsKeyboard()
	}
	return sendWithKeyboard(bot, text, chatID, keyboard)
}

func sendWithKeyboard(bot telegramClient, text string, chatID int64, keyboard interface{}) int {
	// set typing action
	action := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	bot.Send(action)
	lastMessageID := 0

	for _, chunk := range splitStringToChunks(text) {

		// if msgRuneCount < 4096, send it normally
		msg := tgbotapi.NewMessage(chatID, chunk)
		msg.DisableWebPagePreview = true
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = keyboard

		resp, err := bot.Send(msg)
		if err != nil {
			log.Printf("[ERROR] Send: %s", err)
		}
		lastMessageID = resp.MessageID
	}
	return lastMessageID
}

func sendTorrents(bot telegramClient, ud messageWrapper, torrents transmission.Torrents) {
	buf := new(bytes.Buffer)
	for _, torrent := range torrents {
		buf.WriteString(fmt.Sprintf("*%d* `%s` _%s_\n", torrent.ID, ellipsisString(torrent.Name, 25), torrent.TorrentStatus()))
	}

	if buf.Len() == 0 {
		send(bot, "No torrents", ud.Message.Chat.ID, true)
		return
	}

	send(bot, buf.String(), ud.Message.Chat.ID, true)
}

func sendFilteredTorrets(bot telegramClient, client transmissionClient, ud messageWrapper, filter torrentFilter) {
	torrents, err := client.GetTorrents()
	if err != nil {
		send(bot, "Torrents obtain error: "+err.Error(), ud.Message.Chat.ID, true)
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

func invokeError(any interface{}, name string, args ...interface{}) error {
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	err := reflect.ValueOf(any).MethodByName(name).Call(inputs)[0].Interface()
	if err != nil {
		return err.(error)
	}
	return nil
}

func invokeStatus(any interface{}, name string, args ...interface{}) (string, error) {
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	results := reflect.ValueOf(any).MethodByName(name).Call(inputs)
	log.Println(results)
	status := results[0].String()
	err := results[1].Interface()
	if err != nil {
		return status, err.(error)
	}
	return status, nil
}

func progressString(persentage float64, length int) string {
	fill := int(persentage * float64(length))
	empty := length - fill
	return fmt.Sprintf("%s%s", strings.Repeat("█", fill), strings.Repeat("░", empty))
}

func progressBar(t *transmission.Torrent) string {
	return fmt.Sprintf("%s %.1f%% %s ↓%s", progressString(t.PercentDone, 10), t.PercentDone, t.ETA(), humanize.Bytes(t.RateDownload))
}

func splitStringToChunks(text string) []string {
	sub := ""
	subs := []string{}

	runes := bytes.Runes([]byte(text))
	l := len(runes)
	for i, r := range runes {
		sub = sub + string(r)
		if (i+1)%4096 == 0 {
			subs = append(subs, sub)
			sub = ""
		} else if (i + 1) == l {
			subs = append(subs, sub)
		}
	}

	return subs
}
