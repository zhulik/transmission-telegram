package main

import (
	"fmt"
	"github.com/pyed/transmission"
	"log"
	"time"
)

func findFinished(before transmission.Torrents, after transmission.Torrents) (result transmission.Torrents) {
	for _, aT := range after {
		for _, aB := range after {
			if aT.ID == aB.ID && aT.IsFinished != aB.IsFinished {
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

func notifyFinished(bot TelegramClient, client TransmissionClient, ud MessageWrapper) {
	var torrents transmission.Torrents
	send(bot, "I will notify you about finished torrents", ud.Chat.ID)
	for {
		log.Println("Looking for finished torrents")
		newTorrents, err := client.GetTorrents()
		if err != nil {
			log.Println("GetTorrents failed:", err.Error())
			continue
		}

		finished := findFinished(torrents, newTorrents)
		log.Println("Found finished torrents", len(finished))
		if len(finished) > 0 {
			for _, t := range newTorrents {
				go sendFinishedTorrent(bot, t, ud.Chat.ID)
			}
		}
		torrents = newTorrents
		time.Sleep(time.Second * interval)
	}
}
