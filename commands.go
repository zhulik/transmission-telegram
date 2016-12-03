package main

import (
	"fmt"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"strconv"
	"unicode/utf8"
)

// receiveTorrent gets an update that potentially has a .torrent file to add
func receiveTorrent(ud tgbotapi.Update) {
	if ud.Message.Document.FileID == "" {
		return // has no document
	}

	// get the file ID and make the config
	fconfig := tgbotapi.FileConfig{
		FileID: ud.Message.Document.FileID,
	}
	file, err := Bot.GetFile(fconfig)
	if err != nil {
		send("receiver: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	// add by file URL
	add(ud, []string{file.Link(BotToken)})
}

// stop takes id[s] of torrent[s] or 'all' to stop them
func stop(ud tgbotapi.Update, tokens []string) {
	// make sure that we got at least one argument
	if len(tokens) == 0 {
		send("stop: needs an argument", ud.Message.Chat.ID, false)
		return
	}

	// if the first argument is 'all' then stop all torrents
	if tokens[0] == "all" {
		if err := Client.StopAll(); err != nil {
			send("stop: error occurred while stopping some torrents", ud.Message.Chat.ID, false)
			return
		}
		send("stopped all torrents", ud.Message.Chat.ID, false)
		return
	}

	for _, id := range tokens {
		num, err := strconv.Atoi(id)
		if err != nil {
			send(fmt.Sprintf("stop: %s is not a number", id), ud.Message.Chat.ID, false)
			continue
		}
		status, err := Client.StopTorrent(num)
		if err != nil {
			send("stop: "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		torrent, err := Client.GetTorrent(num)
		if err != nil {
			send(fmt.Sprintf("[fail] stop: No torrent with an ID of %d", num), ud.Message.Chat.ID, false)
			return
		}
		send(fmt.Sprintf("[%s] stop: %s", status, torrent.Name), ud.Message.Chat.ID, false)
	}
}

// start takes id[s] of torrent[s] or 'all' to start them
func start(ud tgbotapi.Update, tokens []string) {
	// make sure that we got at least one argument
	if len(tokens) == 0 {
		send("start: needs an argument", ud.Message.Chat.ID, false)
		return
	}

	// if the first argument is 'all' then start all torrents
	if tokens[0] == "all" {
		if err := Client.StartAll(); err != nil {
			send("start: error occurred while starting some torrents", ud.Message.Chat.ID, false)
			return
		}
		send("started all torrents", ud.Message.Chat.ID, false)
		return

	}

	for _, id := range tokens {
		num, err := strconv.Atoi(id)
		if err != nil {
			send(fmt.Sprintf("start: %s is not a number", id), ud.Message.Chat.ID, false)
			continue
		}
		status, err := Client.StartTorrent(num)
		if err != nil {
			send("stop: "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		torrent, err := Client.GetTorrent(num)
		if err != nil {
			send(fmt.Sprintf("[fail] start: No torrent with an ID of %d", num), ud.Message.Chat.ID, false)
			return
		}
		send(fmt.Sprintf("[%s] start: %s", status, torrent.Name), ud.Message.Chat.ID, false)
	}
}

// check takes id[s] of torrent[s] or 'all' to verify them
func check(ud tgbotapi.Update, tokens []string) {
	// make sure that we got at least one argument
	if len(tokens) == 0 {
		send("check: needs an argument", ud.Message.Chat.ID, false)
		return
	}

	// if the first argument is 'all' then start all torrents
	if tokens[0] == "all" {
		if err := Client.VerifyAll(); err != nil {
			send("check: error occurred while verifying some torrents", ud.Message.Chat.ID, false)
			return
		}
		send("verifying all torrents", ud.Message.Chat.ID, false)
		return

	}

	for _, id := range tokens {
		num, err := strconv.Atoi(id)
		if err != nil {
			send(fmt.Sprintf("check: %s is not a number", id), ud.Message.Chat.ID, false)
			continue
		}
		status, err := Client.VerifyTorrent(num)
		if err != nil {
			send("stop: "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		torrent, err := Client.GetTorrent(num)
		if err != nil {
			send(fmt.Sprintf("[fail] check: No torrent with an ID of %d", num), ud.Message.Chat.ID, false)
			return
		}
		send(fmt.Sprintf("[%s] check: %s", status, torrent.Name), ud.Message.Chat.ID, false)
	}

}

// del takes an id or more, and delete the corresponding torrent/s
func del(ud tgbotapi.Update, tokens []string) {
	// make sure that we got an argument
	if len(tokens) == 0 {
		send("del: needs an ID", ud.Message.Chat.ID, false)
		return
	}

	// loop over tokens to read each potential id
	for _, id := range tokens {
		num, err := strconv.Atoi(id)
		if err != nil {
			send(fmt.Sprintf("del: %s is not an ID", id), ud.Message.Chat.ID, false)
			return
		}

		name, err := Client.DeleteTorrent(num, false)
		if err != nil {
			send("del: "+err.Error(), ud.Message.Chat.ID, false)
			return
		}

		send("Deleted: "+name, ud.Message.Chat.ID, false)
	}
}

// deldata takes an id or more, and delete the corresponding torrent/s with their data
func deldata(ud tgbotapi.Update, tokens []string) {
	// make sure that we got an argument
	if len(tokens) == 0 {
		send("deldata: needs an ID", ud.Message.Chat.ID, false)
		return
	}
	// loop over tokens to read each potential id
	for _, id := range tokens {
		num, err := strconv.Atoi(id)
		if err != nil {
			send(fmt.Sprintf("deldata: %s is not an ID", id), ud.Message.Chat.ID, false)
			return
		}

		name, err := Client.DeleteTorrent(num, true)
		if err != nil {
			send("deldata: "+err.Error(), ud.Message.Chat.ID, false)
			return
		}

		send("Deleted with data: "+name, ud.Message.Chat.ID, false)
	}
}

// version sends transmission version + transmission-telegram version
func version(ud tgbotapi.Update) {
	send(fmt.Sprintf("Transmission *%s*\nTransmission-telegram *%s*", Client.Version(), VERSION), ud.Message.Chat.ID, true)
}

// send takes a chat id and a message to send, returns the message id of the send message
func send(text string, chatID int64, markdown bool) int {
	// set typing action
	action := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	Bot.Send(action)

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
		if markdown {
			msg.ParseMode = tgbotapi.ModeMarkdown
		}

		// send current chunk
		if _, err := Bot.Send(msg); err != nil {
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
	if markdown {
		msg.ParseMode = tgbotapi.ModeMarkdown
	}

	resp, err := Bot.Send(msg)
	if err != nil {
		log.Printf("[ERROR] Send: %s", err)
	}

	return resp.MessageID
}
