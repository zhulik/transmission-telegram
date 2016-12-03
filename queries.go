package main

import (
	"bytes"
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
	mdReplacer = strings.NewReplacer("*", "•",
		"[", "(",
		"]", ")",
		"_", "-",
		"`", "'")
)

// list will form and send a list of all the torrents
func list(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	torrents, err := client.GetTorrents()
	if err != nil {
		send(bot, "list: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for _, torrent := range torrents {
		buf.WriteString(fmt.Sprintf("*%d* `%s` _%s_\n", torrent.ID, ellipsisString(torrent.Name, 25), torrent.TorrentStatus()))
	}

	if buf.Len() == 0 {
		send(bot, "list: No torrents", ud.Message.Chat.ID, false)
		return
	}

	send(bot, buf.String(), ud.Message.Chat.ID, true)
}

// downs will send the names of torrents with status 'Downloading' or in queue to
func downs(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	torrents, err := client.GetTorrents()
	if err != nil {
		send(bot, "downs: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for i := range torrents {
		// Downloading or in queue to download
		if torrents[i].Status == transmission.StatusDownloading ||
			torrents[i].Status == transmission.StatusDownloadPending {
			buf.WriteString(fmt.Sprintf("<%d> %s\n", torrents[i].ID, torrents[i].Name))
		}
	}

	if buf.Len() == 0 {
		send(bot, "No downloads", ud.Message.Chat.ID, false)
		return
	}
	send(bot, buf.String(), ud.Message.Chat.ID, false)
}

// seeding will send the names of the torrents with the status 'Seeding' or in the queue to
func seeding(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	torrents, err := client.GetTorrents()
	if err != nil {
		send(bot, "seeding: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for i := range torrents {
		if torrents[i].Status == transmission.StatusSeeding ||
			torrents[i].Status == transmission.StatusSeedPending {
			buf.WriteString(fmt.Sprintf("<%d> %s\n", torrents[i].ID, torrents[i].Name))
		}
	}

	if buf.Len() == 0 {
		send(bot, "No torrents seeding", ud.Message.Chat.ID, false)
		return
	}

	send(bot, buf.String(), ud.Message.Chat.ID, false)

}

// paused will send the names of the torrents with status 'Paused'
func paused(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	torrents, err := client.GetTorrents()
	if err != nil {
		send(bot, "paused: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for i := range torrents {
		if torrents[i].Status == transmission.StatusStopped {
			buf.WriteString(fmt.Sprintf("<%d> %s\n%s (%.1f%%) DL: %s UL: %s  R: %s\n\n",
				torrents[i].ID, torrents[i].Name, torrents[i].TorrentStatus(),
				torrents[i].PercentDone*100, humanize.Bytes(torrents[i].DownloadedEver),
				humanize.Bytes(torrents[i].UploadedEver), torrents[i].Ratio()))
		}
	}

	if buf.Len() == 0 {
		send(bot, "No paused torrents", ud.Message.Chat.ID, false)
		return
	}

	send(bot, buf.String(), ud.Message.Chat.ID, false)
}

// checking will send the names of torrents with the status 'verifying' or in the queue to
func checking(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	torrents, err := client.GetTorrents()
	if err != nil {
		send(bot, "checking: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for i := range torrents {
		if torrents[i].Status == transmission.StatusChecking ||
			torrents[i].Status == transmission.StatusCheckPending {
			buf.WriteString(fmt.Sprintf("<%d> %s\n%s (%.1f%%)\n\n",
				torrents[i].ID, torrents[i].Name, torrents[i].TorrentStatus(),
				torrents[i].PercentDone*100))

		}
	}

	if buf.Len() == 0 {
		send(bot, "No torrents verifying", ud.Message.Chat.ID, false)
		return
	}

	send(bot, buf.String(), ud.Message.Chat.ID, false)
}

// errors will send torrents with errors
func errors(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	torrents, err := client.GetTorrents()
	if err != nil {
		send(bot, "errors: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for i := range torrents {
		if torrents[i].Error != 0 {
			buf.WriteString(fmt.Sprintf("<%d> %s\n%s\n",
				torrents[i].ID, torrents[i].Name, torrents[i].ErrorString))
		}
	}
	if buf.Len() == 0 {
		send(bot, "No errors", ud.Message.Chat.ID, false)
		return
	}
	send(bot, buf.String(), ud.Message.Chat.ID, false)
}

// sort changes torrents sorting
func sort(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	if len(ud.Tokens()) == 0 {
		send(bot, `sort takes one of:
			(*id, name, age, size, progress, downspeed, upspeed, download, upload, ratio*)
			optionally start with (*rev*) for reversed order
			e.g. "*sort rev size*" to get biggest torrents first.`, ud.Message.Chat.ID, true)
		return
	}

	var reversed bool
	tokens := ud.Tokens()
	if strings.ToLower(ud.Tokens()[0]) == "rev" {
		reversed = true
		tokens = ud.Tokens()[1:]
	}

	switch strings.ToLower(tokens[0]) {
	case "id":
		if reversed {
			client.SetSort(transmission.SortRevID)
			break
		}
		client.SetSort(transmission.SortID)
	case "name":
		if reversed {
			client.SetSort(transmission.SortRevName)
			break
		}
		client.SetSort(transmission.SortName)
	case "age":
		if reversed {
			client.SetSort(transmission.SortRevAge)
			break
		}
		client.SetSort(transmission.SortAge)
	case "size":
		if reversed {
			client.SetSort(transmission.SortRevSize)
			break
		}
		client.SetSort(transmission.SortSize)
	case "progress":
		if reversed {
			client.SetSort(transmission.SortRevProgress)
			break
		}
		client.SetSort(transmission.SortProgress)
	case "downspeed":
		if reversed {
			client.SetSort(transmission.SortRevDownSpeed)
			break
		}
		client.SetSort(transmission.SortDownSpeed)
	case "upspeed":
		if reversed {
			client.SetSort(transmission.SortRevUpSpeed)
			break
		}
		client.SetSort(transmission.SortUpSpeed)
	case "download":
		if reversed {
			client.SetSort(transmission.SortRevDownloaded)
			break
		}
		client.SetSort(transmission.SortDownloaded)
	case "upload":
		if reversed {
			client.SetSort(transmission.SortRevUploaded)
			break
		}
		client.SetSort(transmission.SortUploaded)
	case "ratio":
		if reversed {
			client.SetSort(transmission.SortRevRatio)
			break
		}
		client.SetSort(transmission.SortRatio)
	default:
		send(bot, "unkown sorting method", ud.Message.Chat.ID, false)
		return
	}

	if reversed {
		send(bot, "sort: reversed "+tokens[0], ud.Message.Chat.ID, false)
		return
	}
	send(bot, "sort: "+tokens[0], ud.Message.Chat.ID, false)
}

var trackerRegex = regexp.MustCompile(`[https?|udp]://([^:/]*)`)

// trackers will send a list of trackers and how many torrents each one has
func trackers(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	torrents, err := client.GetTorrents()
	if err != nil {
		send(bot, "trackers: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	trackers := make(map[string]int)

	for i := range torrents {
		for _, tracker := range torrents[i].Trackers {
			sm := trackerRegex.FindSubmatch([]byte(tracker.Announce))
			if len(sm) > 1 {
				currentTracker := string(sm[1])
				n, ok := trackers[currentTracker]
				if !ok {
					trackers[currentTracker] = 1
					continue
				}
				trackers[currentTracker] = n + 1
			}
		}
	}

	buf := new(bytes.Buffer)
	for k, v := range trackers {
		buf.WriteString(fmt.Sprintf("%d - %s\n", v, k))
	}

	if buf.Len() == 0 {
		send(bot, "No trackers!", ud.Message.Chat.ID, false)
		return
	}
	send(bot, buf.String(), ud.Message.Chat.ID, false)
}

// count returns current torrents count per status
func count(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	torrents, err := client.GetTorrents()
	if err != nil {
		send(bot, "count: "+err.Error(), ud.Message.Chat.ID, false)
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

	msg := fmt.Sprintf("Downloading: %d\nSeeding: %d\nPaused: %d\nVerifying: %d\n\n- Waiting to -\nDownload: %d\nSeed: %d\nVerify: %d\n\nTotal: %d",
		downloading, seeding, stopped, checking, downloadingQ, seedingQ, checkingQ, len(torrents))

	send(bot, msg, ud.Message.Chat.ID, false)

}

// search takes a query and returns torrents with match
func search(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	// make sure that we got a query
	if len(ud.Tokens()) == 0 {
		send(bot, "search: needs an argument", ud.Message.Chat.ID, false)
		return
	}

	query := strings.Join(ud.Tokens(), " ")
	// "(?i)" for case insensitivity
	regx, err := regexp.Compile("(?i)" + query)
	if err != nil {
		send(bot, "search: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	torrents, err := client.GetTorrents()
	if err != nil {
		send(bot, "search: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for i := range torrents {
		if regx.MatchString(torrents[i].Name) {
			buf.WriteString(fmt.Sprintf("<%d> %s\n", torrents[i].ID, torrents[i].Name))
		}
	}
	if buf.Len() == 0 {
		send(bot, "No matches!", ud.Message.Chat.ID, false)
		return
	}
	send(bot, buf.String(), ud.Message.Chat.ID, false)
}

// info takes an id of a torrent and returns some info about it
func info(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	if len(ud.Tokens()) == 0 {
		send(bot, "info: needs a torrent ID number", ud.Message.Chat.ID, false)
		return
	}

	for _, id := range ud.Tokens() {
		torrentID, err := strconv.Atoi(id)
		if err != nil {
			send(bot, fmt.Sprintf("info: %s is not a number", id), ud.Message.Chat.ID, false)
			continue
		}

		// get the torrent
		torrent, err := client.GetTorrent(torrentID)
		if err != nil {
			send(bot, fmt.Sprintf("info: Can't find a torrent with an ID of %d", torrentID), ud.Message.Chat.ID, false)
			continue
		}

		// get the trackers using 'trackerRegex'
		var trackers string
		for _, tracker := range torrent.Trackers {
			sm := trackerRegex.FindSubmatch([]byte(tracker.Announce))
			if len(sm) > 1 {
				trackers += string(sm[1]) + " "
			}
		}

		// format the info
		torrentName := mdReplacer.Replace(torrent.Name) // escape markdown
		info := fmt.Sprintf("`<%d>` *%s*\n%s *%s* of *%s* (*%.1f%%*) ↓ *%s*  ↑ *%s* R: *%s*\nDL: *%s* UP: *%s*\nAdded: *%s*, ETA: *%s*\nTrackers: `%s`",
			torrent.ID, torrentName, torrent.TorrentStatus(), humanize.Bytes(torrent.Have()), humanize.Bytes(torrent.SizeWhenDone),
			torrent.PercentDone*100, humanize.Bytes(torrent.RateDownload), humanize.Bytes(torrent.RateUpload), torrent.Ratio(),
			humanize.Bytes(torrent.DownloadedEver), humanize.Bytes(torrent.UploadedEver), time.Unix(torrent.AddedDate, 0).Format(time.Stamp),
			torrent.ETA(), trackers)

		// send it
		msgID := send(bot, info, ud.Message.Chat.ID, true)

		// this go-routine will make the info live for 'duration * interval'
		// takes trackers so we don't have to regex them over and over.
		go func(trackers string, torrentID, msgID int) {
			for i := 0; i < duration; i++ {
				time.Sleep(time.Second * interval)
				torrent, err := client.GetTorrent(torrentID)
				if err != nil {
					continue // skip this iteration if there's an error retrieving the torrent's info
				}

				torrentName := mdReplacer.Replace(torrent.Name)
				info := fmt.Sprintf("`<%d>` *%s*\n%s *%s* of *%s* (*%.1f%%*) ↓ *%s*  ↑ *%s* R: *%s*\nDL: *%s* UP: *%s*\nAdded: *%s*, ETA: *%s*\nTrackers: `%s`",
					torrent.ID, torrentName, torrent.TorrentStatus(), humanize.Bytes(torrent.Have()), humanize.Bytes(torrent.SizeWhenDone),
					torrent.PercentDone*100, humanize.Bytes(torrent.RateDownload), humanize.Bytes(torrent.RateUpload), torrent.Ratio(),
					humanize.Bytes(torrent.DownloadedEver), humanize.Bytes(torrent.UploadedEver), time.Unix(torrent.AddedDate, 0).Format(time.Stamp),
					torrent.ETA(), trackers)

				// update the message
				editConf := tgbotapi.NewEditMessageText(ud.Message.Chat.ID, msgID, info)
				editConf.ParseMode = tgbotapi.ModeMarkdown
				bot.Send(editConf)

			}

			// at the end write dashes to indicate that we are done being live.
			torrentName := mdReplacer.Replace(torrent.Name)
			info := fmt.Sprintf("`<%d>` *%s*\n%s *%s* of *%s* (*%.1f%%*) ↓ *- B*  ↑ *- B* R: *%s*\nDL: *%s* UP: *%s*\nAdded: *%s*, ETA: *-*\nTrackers: `%s`",
				torrent.ID, torrentName, torrent.TorrentStatus(), humanize.Bytes(torrent.Have()), humanize.Bytes(torrent.SizeWhenDone),
				torrent.PercentDone*100, torrent.Ratio(), humanize.Bytes(torrent.DownloadedEver), humanize.Bytes(torrent.UploadedEver),
				time.Unix(torrent.AddedDate, 0).Format(time.Stamp), trackers)

			editConf := tgbotapi.NewEditMessageText(ud.Message.Chat.ID, msgID, info)
			editConf.ParseMode = tgbotapi.ModeMarkdown
			bot.Send(editConf)
		}(trackers, torrentID, msgID)
	}
}

// stats echo back transmission stats
func stats(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	stats, err := client.GetStats()
	if err != nil {
		send(bot, "stats: "+err.Error(), ud.Message.Chat.ID, false)
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

	send(bot, msg, ud.Message.Chat.ID, true)
}

// speed will echo back the current download and upload speeds
func speed(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud UpdateWrapper) {
	// keep track of the returned message ID from 'send()' to edit the message.
	var msgID int
	for i := 0; i < duration; i++ {
		stats, err := client.GetStats()
		if err != nil {
			send(bot, "speed: "+err.Error(), ud.Message.Chat.ID, false)
			return
		}

		msg := fmt.Sprintf("↓ %s  ↑ %s", humanize.Bytes(stats.DownloadSpeed), humanize.Bytes(stats.UploadSpeed))

		// if we haven't send a message, send it and save the message ID to edit it the next iteration
		if msgID == 0 {
			msgID = send(bot, msg, ud.Message.Chat.ID, false)
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
