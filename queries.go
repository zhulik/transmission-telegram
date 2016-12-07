package main

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/pyed/transmission"
	"gopkg.in/telegram-bot-api.v4"
	"regexp"
	"strconv"
	"strings"
)

var (
	trackerRegex = regexp.MustCompile(`[https?|udp]://([^:/]*)`)
)

// list will form and send a list of all the torrents
func list(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud MessageWrapper) {
	if len(ud.Tokens()) > 0 {
		mode := ud.Tokens()[0]
		switch mode {
		case "dl":
			downs(bot, client, ud)
		case "sd":
			seeding(bot, client, ud)
		case "pa":
			paused(bot, client, ud)
		case "ch":
			checking(bot, client, ud)
		case "er":
			errors(bot, client, ud)
		}
	} else {
		sendFilteredTorrets(bot, client, ud, func(t *transmission.Torrent) bool { return true })
	}
}

// downs will send the names of torrents with status 'Downloading' or in queue to
func downs(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud MessageWrapper) {
	sendFilteredTorrets(bot, client, ud, func(t *transmission.Torrent) bool {
		return t.Status == transmission.StatusDownloading ||
			t.Status == transmission.StatusDownloadPending
	})
}

// seeding will send the names of the torrents with the status 'Seeding' or in the queue to
func seeding(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud MessageWrapper) {
	sendFilteredTorrets(bot, client, ud, func(t *transmission.Torrent) bool {
		return t.Status == transmission.StatusSeeding ||
			t.Status == transmission.StatusSeedPending
	})
}

// paused will send the names of the torrents with status 'Paused'
func paused(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud MessageWrapper) {
	sendFilteredTorrets(bot, client, ud, func(t *transmission.Torrent) bool {
		return t.Status == transmission.StatusStopped
	})
}

// checking will send the names of torrents with the status 'verifying' or in the queue to
func checking(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud MessageWrapper) {
	sendFilteredTorrets(bot, client, ud, func(t *transmission.Torrent) bool {
		return t.Status == transmission.StatusChecking ||
			t.Status == transmission.StatusCheckPending
	})
}

// errors will send torrents with errors
func errors(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud MessageWrapper) {
	sendFilteredTorrets(bot, client, ud, func(t *transmission.Torrent) bool {
		return t.Error != 0
	})
}

// search takes a query and returns torrents with match
func search(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud MessageWrapper) {
	// make sure that we got a query
	if len(ud.Tokens()) == 0 {
		send(bot, "*search*: needs an argument", ud.Chat.ID)
		return
	}

	query := strings.Join(ud.Tokens(), " ")
	// "(?i)" for case insensitivity
	regx, err := regexp.Compile("(?i)" + query)
	if err != nil {
		send(bot, fmt.Sprintf("*search*: `%s`", err.Error()), ud.Chat.ID)
		return
	}

	sendFilteredTorrets(bot, client, ud, func(t *transmission.Torrent) bool {
		return regx.MatchString(t.Name)
	})
}

// count returns current torrents count per status
func count(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud MessageWrapper) {
	torrents, err := client.GetTorrents()
	if err != nil {
		send(bot, fmt.Sprintf("*count*: `%s`", err.Error()), ud.Chat.ID)
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

	send(bot, msg, ud.Chat.ID)

}

// info takes an id of a torrent and returns some info about it
func info(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud MessageWrapper) {
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

// stats echo back transmission stats
func stats(bot *tgbotapi.BotAPI, client *transmission.TransmissionClient, ud MessageWrapper) {
	stats, err := client.GetStats()
	if err != nil {
		send(bot, fmt.Sprintf("*stats*: `%s`", err.Error()), ud.Chat.ID)
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

	send(bot, msg, ud.Chat.ID)
}
