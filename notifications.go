package main

import (
	"fmt"
	"github.com/pyed/transmission"
	"github.com/zhulik/transmission-telegram/settings"
	"log"
	"time"
)

func findFinished(before transmission.Torrents, after transmission.Torrents) (result transmission.Torrents) {
	for _, aT := range after {
		for _, bT := range before {
			if aT.ID == bT.ID && (bT.Status == transmission.StatusDownloading && (aT.Status == transmission.StatusSeeding || aT.Status == transmission.StatusSeedPending)) {
				result = append(result, aT)
			}
		}
	}
	return
}

func sendFinishedTorrent(bot TelegramClient, t *transmission.Torrent, chatID int64) {
	msg := fmt.Sprintf("*%d* `%s` is finished!", t.ID, ellipsisString(t.Name, 25))
	send(bot, msg, chatID)
	log.Println("Finished torrent was sent")
}

func notifyFinished(bot TelegramClient, client TransmissionClient, ud MessageWrapper, s settings.Settings) {
	var torrents transmission.Torrents
	send(bot, "I will notify you about finished torrents", ud.Chat.ID)
	for {
		newTorrents, err := client.GetTorrents()
		if err != nil {
			log.Println("GetTorrents failed:", err.Error())
			continue
		}

		for _, t := range findFinished(torrents, newTorrents) {
			go sendFinishedTorrent(bot, t, ud.Chat.ID)
		}
		torrents = newTorrents
		time.Sleep(time.Second * interval)
	}
}
