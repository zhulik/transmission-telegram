package main

import (
	"bytes"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/pyed/transmission"
	"gopkg.in/telegram-bot-api.v4"
	"strconv"
	"time"
)

const (
	duration               = 60
	interval time.Duration = 2
)

// info takes an id of a torrent and returns some info about it
func info(bot TelegramClient, client TransmissionClient, ud MessageWrapper) {
	if len(ud.Tokens()) == 0 {
		send(bot, "*info*: needs a torrent ID number", ud.Chat.ID)
		return
	}

	for _, id := range ud.Tokens() {
		torrentID, err := strconv.Atoi(id)
		if err != nil {
			send(bot, fmt.Sprintf("*info*: %s is not a number", id), ud.Chat.ID)
			continue
		}

		_, err = client.GetTorrent(torrentID)
		if err != nil {
			send(bot, fmt.Sprintf("*info*: Can't find a torrent with an ID of %d", torrentID), ud.Chat.ID)
			continue
		}
		go updateTorrentInfo(bot, client, ud, torrentID)
	}
}

func updateTorrentInfo(bot TelegramClient, client TransmissionClient, ud MessageWrapper, torrentID int) {
	msgID := -1
	for i := 0; i < duration; i++ {
		torrent, err := client.GetTorrent(torrentID)
		if err != nil {
			continue // skip this iteration if there's an error retrieving the torrent's info
		}

		info := fmt.Sprintf("*%d* `%s`\n%s *%s* of *%s* (*%.1f%%*) ↓ *%s*  ↑ *%s* R: *%s*\nDL: *%s* UP: *%s*\nAdded: *%s*, ETA: *%s*",
			torrent.ID, torrent.Name, torrent.TorrentStatus(), humanize.Bytes(torrent.Have()), humanize.Bytes(torrent.SizeWhenDone),
			torrent.PercentDone*100, humanize.Bytes(torrent.RateDownload), humanize.Bytes(torrent.RateUpload), torrent.Ratio(),
			humanize.Bytes(torrent.DownloadedEver), humanize.Bytes(torrent.UploadedEver), time.Unix(torrent.AddedDate, 0).Format(time.Stamp),
			torrent.ETA())

		// update the message
		if msgID == -1 {
			msgID = send(bot, info, ud.Chat.ID)
			time.Sleep(time.Second * interval)
			continue
		} else {
			editConf := tgbotapi.NewEditMessageText(ud.Chat.ID, msgID, info)
			editConf.ParseMode = tgbotapi.ModeMarkdown
			bot.Send(editConf)
		}
		time.Sleep(time.Second * interval)
	}

	torrent, err := client.GetTorrent(torrentID)
	if err != nil {
		return
	}

	// at the end write dashes to indicate that we are done being live.
	info := fmt.Sprintf("*%d* `%s`\n%s *%s* of *%s* (*%.1f%%*) ↓ *- B*  ↑ *- B* R: *%s*\nDL: *%s* UP: *%s*\nAdded: *%s*, ETA: *-*\n",
		torrent.ID, torrent.Name, torrent.TorrentStatus(), humanize.Bytes(torrent.Have()), humanize.Bytes(torrent.SizeWhenDone),
		torrent.PercentDone*100, torrent.Ratio(), humanize.Bytes(torrent.DownloadedEver), humanize.Bytes(torrent.UploadedEver),
		time.Unix(torrent.AddedDate, 0).Format(time.Stamp))

	editConf := tgbotapi.NewEditMessageText(ud.Chat.ID, msgID, info)
	editConf.ParseMode = tgbotapi.ModeMarkdown
	bot.Send(editConf)
}

// speed will echo back the current download and upload speeds
func speed(bot TelegramClient, client TransmissionClient, ud MessageWrapper) {
	// keep track of the returned message ID from 'send()' to edit the message.
	msgID := -1
	for i := 0; i < duration; i++ {
		stats, err := client.GetStats()
		if err != nil {
			send(bot, fmt.Sprintf("*speed*: `%s`", err.Error()), ud.Chat.ID)
			return
		}

		msg := fmt.Sprintf("↓ *%s*  ↑ *%s*", humanize.Bytes(stats.DownloadSpeed), humanize.Bytes(stats.UploadSpeed))

		// if we haven't send a message, send it and save the message ID to edit it the next iteration
		if msgID == -1 {
			msgID = send(bot, msg, ud.Chat.ID)
			time.Sleep(time.Second * interval)
			continue
		}

		// we have sent the message, let's update.
		editConf := tgbotapi.NewEditMessageText(ud.Chat.ID, msgID, msg)
		bot.Send(editConf)
		editConf.ParseMode = tgbotapi.ModeMarkdown
		time.Sleep(time.Second * interval)
	}

	editConf := tgbotapi.NewEditMessageText(ud.Chat.ID, msgID, "↓ - B  ↑ - B")
	editConf.ParseMode = tgbotapi.ModeMarkdown
	bot.Send(editConf)
}

// progress echo bach progress and other info for downloading torrents
func progress(bot TelegramClient, client TransmissionClient, ud MessageWrapper) {
	msgID := -1
	for i := 0; i < duration; i++ {
		torrents, err := client.GetTorrents()
		if err != nil {
			send(bot, "Torrents obtain error: "+err.Error(), ud.Chat.ID)
			continue
		}

		buf := new(bytes.Buffer)
		for _, t := range torrents {
			if t.Status == transmission.StatusDownloading {
				buf.WriteString(fmt.Sprintf("*%d* `%s`\n%s %.1f%% %s ↓%s\n", t.ID, ellipsisString(t.Name, 30), progressString(t.PercentDone, 10), t.PercentDone*100, t.ETA(), humanize.Bytes(t.RateDownload)))
			}
		}

		if buf.Len() == 0 {
			send(bot, "No torrents", ud.Chat.ID)
			return
		}

		if msgID == -1 {
			msgID = send(bot, buf.String(), ud.Chat.ID)
			continue
		}

		editConf := tgbotapi.NewEditMessageText(ud.Chat.ID, msgID, buf.String())
		editConf.ParseMode = tgbotapi.ModeMarkdown
		bot.Send(editConf)
		time.Sleep(time.Second * interval)
	}
}
