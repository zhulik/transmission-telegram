package main

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/pyed/transmission"
	"gopkg.in/telegram-bot-api.v4"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	trackerRegex = regexp.MustCompile(`[https?|udp]://([^:/]*)`)
)

// list will form and send a list of all the torrents
func list(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	sendFilteredTorrets(bot, client, ud, func(t *transmission.Torrent) bool { return true })
}

// downs will send the names of torrents with status 'Downloading' or in queue to
func downs(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	sendFilteredTorrets(bot, client, ud, func(t *transmission.Torrent) bool {
		return t.Status == transmission.StatusDownloading ||
			t.Status == transmission.StatusDownloadPending
	})
}

// seeding will send the names of the torrents with the status 'Seeding' or in the queue to
func seeding(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	sendFilteredTorrets(bot, client, ud, func(t *transmission.Torrent) bool {
		return t.Status == transmission.StatusSeeding ||
			t.Status == transmission.StatusSeedPending
	})
}

// paused will send the names of the torrents with status 'Paused'
func paused(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	sendFilteredTorrets(bot, client, ud, func(t *transmission.Torrent) bool {
		return t.Status == transmission.StatusStopped
	})
}

// checking will send the names of torrents with the status 'verifying' or in the queue to
func checking(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	sendFilteredTorrets(bot, client, ud, func(t *transmission.Torrent) bool {
		return t.Status == transmission.StatusChecking ||
			t.Status == transmission.StatusCheckPending
	})
}

// errors will send torrents with errors
func errors(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	sendFilteredTorrets(bot, client, ud, func(t *transmission.Torrent) bool {
		return t.Error != 0
	})
}

// search takes a query and returns torrents with match
func search(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	// make sure that we got a query
	if len(ud.Tokens()) == 0 {
		send(bot, "*search*: needs an argument", ud.Message.Chat.ID)
		return
	}

	query := strings.Join(ud.Tokens(), " ")
	// "(?i)" for case insensitivity
	regx, err := regexp.Compile("(?i)" + query)
	if err != nil {
		send(bot, "*search*: "+err.Error(), ud.Message.Chat.ID)
		return
	}

	sendFilteredTorrets(bot, client, ud, func(t *transmission.Torrent) bool {
		return regx.MatchString(t.Name)
	})
}

// count returns current torrents count per status
func count(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	torrents, err := client.GetTorrents()
	if err != nil {
		send(bot, "*count*: "+err.Error(), ud.Message.Chat.ID)
		return
	}

	var downloading, seeding, stopped, checking, downloadingQ, seedingQ, checkingQ int

	for i := range torrents {
		switch torrents[i].Status {
		case transmission.StatusDownloading:
			downloading++
		case transmission.StatusSeeding:
			seeding++
		case transmission.StatusStopped:
			stopped++
		case transmission.StatusChecking:
			checking++
		case transmission.StatusDownloadPending:
			downloadingQ++
		case transmission.StatusSeedPending:
			seedingQ++
		case transmission.StatusCheckPending:
			checkingQ++
		}
	}

	msg := fmt.Sprintf("*Downloading*: %d\n*Seeding*: %d\n*Paused*: %d\n*Verifying*: %d\n\n- Waiting to -\n*Download*: %d\n*Seed*: %d\n*Verify*: %d\n\n*Total*: %d",
		downloading, seeding, stopped, checking, downloadingQ, seedingQ, checkingQ, len(torrents))

	send(bot, msg, ud.Message.Chat.ID)

}

// info takes an id of a torrent and returns some info about it
func info(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	if len(ud.Tokens()) == 0 {
		send(bot, "*info*: needs a torrent ID number", ud.Message.Chat.ID)
		return
	}

	for _, id := range ud.Tokens() {
		torrentID, err := strconv.Atoi(id)
		if err != nil {
			send(bot, fmt.Sprintf("*info*: %s is not a number", id), ud.Message.Chat.ID)
			continue
		}

		// get the torrent
		torrent, err := client.GetTorrent(torrentID)
		if err != nil {
			send(bot, fmt.Sprintf("*info*: Can't find a torrent with an ID of %d", torrentID), ud.Message.Chat.ID)
			continue
		}

		var trackers string
		for _, tracker := range torrent.Trackers {
			sm := trackerRegex.FindSubmatch([]byte(tracker.Announce))
			if len(sm) > 1 {
				trackers += string(sm[1]) + " "
			}
		}

		// format the info
		info := fmt.Sprintf("*%d* `%s`\n%s *%s* of *%s* (*%.1f%%*) ↓ *%s*  ↑ *%s* R: *%s*\nDL: *%s* UP: *%s*\nAdded: *%s*, ETA: *%s*\nTrackers: `%s`",
			torrent.ID, torrent.Name, torrent.TorrentStatus(), humanize.Bytes(torrent.Have()), humanize.Bytes(torrent.SizeWhenDone),
			torrent.PercentDone*100, humanize.Bytes(torrent.RateDownload), humanize.Bytes(torrent.RateUpload), torrent.Ratio(),
			humanize.Bytes(torrent.DownloadedEver), humanize.Bytes(torrent.UploadedEver), time.Unix(torrent.AddedDate, 0).Format(time.Stamp),
			torrent.ETA(), trackers)

		// send it
		msgID := send(bot, info, ud.Message.Chat.ID)

		// this go-routine will make the info live for 'duration * interval'
		// takes trackers so we don't have to regex them over and over.
		go updateTorrentInfo(bot, client, ud, trackers, torrentID, msgID)
	}
}

func updateTorrentInfo(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper, trackers string, torrentID, msgID int) {
	for i := 0; i < duration; i++ {
		time.Sleep(time.Second * interval)
		torrent, err := client.GetTorrent(torrentID)
		if err != nil {
			continue // skip this iteration if there's an error retrieving the torrent's info
		}

		info := fmt.Sprintf("*%d* `%s`\n%s *%s* of *%s* (*%.1f%%*) ↓ *%s*  ↑ *%s* R: *%s*\nDL: *%s* UP: *%s*\nAdded: *%s*, ETA: *%s*\nTrackers: `%s`",
			torrent.ID, torrent.Name, torrent.TorrentStatus(), humanize.Bytes(torrent.Have()), humanize.Bytes(torrent.SizeWhenDone),
			torrent.PercentDone*100, humanize.Bytes(torrent.RateDownload), humanize.Bytes(torrent.RateUpload), torrent.Ratio(),
			humanize.Bytes(torrent.DownloadedEver), humanize.Bytes(torrent.UploadedEver), time.Unix(torrent.AddedDate, 0).Format(time.Stamp),
			torrent.ETA(), trackers)

		// update the message
		editConf := tgbotapi.NewEditMessageText(ud.Message.Chat.ID, msgID, info)
		editConf.ParseMode = tgbotapi.ModeMarkdown
		bot.Send(editConf)

	}

	torrent, err := client.GetTorrent(torrentID)
	if err != nil {
		return
	}

	// at the end write dashes to indicate that we are done being live.
	info := fmt.Sprintf("*%d* `%s`\n%s *%s* of *%s* (*%.1f%%*) ↓ *- B*  ↑ *- B* R: *%s*\nDL: *%s* UP: *%s*\nAdded: *%s*, ETA: *-*\nTrackers: `%s`",
		torrent.ID, torrent.Name, torrent.TorrentStatus(), humanize.Bytes(torrent.Have()), humanize.Bytes(torrent.SizeWhenDone),
		torrent.PercentDone*100, torrent.Ratio(), humanize.Bytes(torrent.DownloadedEver), humanize.Bytes(torrent.UploadedEver),
		time.Unix(torrent.AddedDate, 0).Format(time.Stamp), trackers)

	editConf := tgbotapi.NewEditMessageText(ud.Message.Chat.ID, msgID, info)
	editConf.ParseMode = tgbotapi.ModeMarkdown
	bot.Send(editConf)
}

// stats echo back transmission stats
func stats(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	stats, err := client.GetStats()
	if err != nil {
		send(bot, "*stats*: "+err.Error(), ud.Message.Chat.ID)
		return
	}

	msg := fmt.Sprintf(
		`
		Total: *%d*
		Active: *%d*
		Paused: *%d*

		_Current Stats_
		Downloaded: *%s*
		Uploaded: *%s*
		Running time: *%s*

		_Accumulative Stats_
		Sessions: *%d*
		Downloaded: *%s*
		Uploaded: *%s*
		Total Running time: *%s*
		`,

		stats.TorrentCount,
		stats.ActiveTorrentCount,
		stats.PausedTorrentCount,
		humanize.Bytes(stats.CurrentStats.DownloadedBytes),
		humanize.Bytes(stats.CurrentStats.UploadedBytes),
		stats.CurrentActiveTime(),
		stats.CumulativeStats.SessionCount,
		humanize.Bytes(stats.CumulativeStats.DownloadedBytes),
		humanize.Bytes(stats.CumulativeStats.UploadedBytes),
		stats.CumulativeActiveTime(),
	)

	send(bot, msg, ud.Message.Chat.ID)
}

// speed will echo back the current download and upload speeds
func speed(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	// keep track of the returned message ID from 'send()' to edit the message.
	var msgID int
	for i := 0; i < duration; i++ {
		stats, err := client.GetStats()
		if err != nil {
			send(bot, "*speed*: "+err.Error(), ud.Message.Chat.ID)
			return
		}

		msg := fmt.Sprintf("↓ *%s*  ↑ *%s*", humanize.Bytes(stats.DownloadSpeed), humanize.Bytes(stats.UploadSpeed))

		// if we haven't send a message, send it and save the message ID to edit it the next iteration
		if msgID == 0 {
			msgID = send(bot, msg, ud.Message.Chat.ID)
			time.Sleep(time.Second * interval)
			continue
		}

		// we have sent the message, let's update.
		editConf := tgbotapi.NewEditMessageText(ud.Message.Chat.ID, msgID, msg)
		bot.Send(editConf)
		time.Sleep(time.Second * interval)
	}

	// after the 10th iteration, show dashes to indicate that we are done updating.
	editConf := tgbotapi.NewEditMessageText(ud.Message.Chat.ID, msgID, "↓ - B  ↑ - B")
	bot.Send(editConf)
}
