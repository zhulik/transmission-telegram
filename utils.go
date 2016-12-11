package main

import (
	"bytes"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/pyed/transmission"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"reflect"
	"strings"
	"unicode/utf8"
)

var (
	SortingMethods = map[SortMethod]transmission.Sorting{
		SortMethod{"id", false}:       transmission.SortID,
		SortMethod{"id", true}:        transmission.SortRevID,
		SortMethod{"name", false}:     transmission.SortName,
		SortMethod{"name", true}:      transmission.SortRevName,
		SortMethod{"age", false}:      transmission.SortAge,
		SortMethod{"age", true}:       transmission.SortRevAge,
		SortMethod{"size", false}:     transmission.SortSize,
		SortMethod{"size", true}:      transmission.SortRevSize,
		SortMethod{"progress", false}: transmission.SortProgress,
		SortMethod{"progress", true}:  transmission.SortRevProgress,
		SortMethod{"downsped", false}: transmission.SortDownSpeed,
		SortMethod{"downsped", true}:  transmission.SortRevDownSpeed,
		SortMethod{"upspeed", false}:  transmission.SortUpSpeed,
		SortMethod{"upspeed", true}:   transmission.SortRevUpSpeed,

		SortMethod{"download", false}: transmission.SortDownloaded,
		SortMethod{"download", true}:  transmission.SortRevDownloaded,

		SortMethod{"upload", false}: transmission.SortUploaded,
		SortMethod{"upload", true}:  transmission.SortRevUploaded,

		SortMethod{"ratio", false}: transmission.SortRatio,
		SortMethod{"ratio", true}:  transmission.SortRevRatio,
	}
)

type SortMethod struct {
	name     string
	reversed bool
}

type MessageWrapper struct {
	*tgbotapi.Message
	command string
	tokens  []string
}

func WrapMessage(message *tgbotapi.Message) MessageWrapper {
	tokens := strings.Split(message.Text, " ")
	command := strings.ToLower(tokens[0])
	args := tokens[1:]
	return MessageWrapper{message, command, args}
}

func (w MessageWrapper) Command() string {
	return w.command
}

func (w MessageWrapper) Tokens() []string {
	return w.tokens
}

type CommandHandler func(bot TelegramClient, client TransmissionClient, ud MessageWrapper)
type TorrentFilter func(torrent *transmission.Torrent) bool

func ellipsisString(str string, length int) string {
	if utf8.RuneCountInString(str) > length {
		return string([]rune(str)[:length-3]) + "..."
	}
	return str
}

// send takes a chat id and a message to send, returns the message id of the send message
func send(bot TelegramClient, text string, chatID int64) int {
	// set typing action
	action := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	bot.Send(action)

	// check the rune count, telegram is limited to 4096 chars per message;
	// so if our message is > 4096, split it in chunks the send them.
	msgRuneCount := utf8.RuneCountInString(text)
LenCheck:
	stop := 4095
	if msgRuneCount > 4096 {
		for text[stop] != 10 { // '\n'
			stop--
		}
		msg := tgbotapi.NewMessage(chatID, text[:stop])
		msg.DisableWebPagePreview = true
		msg.ParseMode = tgbotapi.ModeMarkdown

		// send current chunk
		if _, err := bot.Send(msg); err != nil {
			log.Printf("[ERROR] Send: %s", err)
		}
		// move to the next chunk
		text = text[stop:]
		msgRuneCount = utf8.RuneCountInString(text)
		goto LenCheck
	}

	// if msgRuneCount < 4096, send it normally
	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableWebPagePreview = true
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = commandsKeyboard()

	resp, err := bot.Send(msg)
	if err != nil {
		log.Printf("[ERROR] Send: %s", err)
	}

	return resp.MessageID
}

func sendTorrents(bot TelegramClient, ud MessageWrapper, torrents transmission.Torrents) {
	buf := new(bytes.Buffer)
	for _, torrent := range torrents {
		buf.WriteString(fmt.Sprintf("*%d* `%s` _%s_\n", torrent.ID, ellipsisString(torrent.Name, 25), torrent.TorrentStatus()))
	}

	if buf.Len() == 0 {
		send(bot, "No torrents", ud.Message.Chat.ID)
		return
	}

	send(bot, buf.String(), ud.Message.Chat.ID)
}

func sendFilteredTorrets(bot TelegramClient, client TransmissionClient, ud MessageWrapper, filter TorrentFilter) {
	torrents, err := client.GetTorrents()
	if err != nil {
		send(bot, "Torrents obtain error: "+err.Error(), ud.Message.Chat.ID)
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

func InvokeError(any interface{}, name string, args ...interface{}) error {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	err := reflect.ValueOf(any).MethodByName(name).Call(inputs)[0].Interface()
	if err != nil {
		return err.(error)
	}
	return nil
}

func InvokeStatus(any interface{}, name string, args ...interface{}) (string, error) {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
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
