package main

import (
	"fmt"
	"github.com/pyed/transmission"
	"github.com/zhulik/transmission-telegram/settings"
	"log"
	"strings"
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

func notifyFinished(bot TelegramClient, client TransmissionClient, masters []string, s settings.Settings) {
	var torrents transmission.Torrents
	for {
		newTorrents, err := client.GetTorrents()
		if err != nil {
			log.Println("GetTorrents failed:", err.Error())
			continue
		}

		for _, t := range findFinished(torrents, newTorrents) {
			for _, master := range masters {
				notify, err := s.GetUserNotification(master)
				if err != nil {
					log.Println("GetUserNotification failed:", err.Error())
					continue
				}
				if notify {
					id, err := s.GetUserID(master)
					if err != nil {
						log.Println("GetUserID failed:", err.Error())
						continue
					}
					go sendFinishedTorrent(bot, t, id)
				}
			}

		}
		torrents = newTorrents
		time.Sleep(time.Second * interval)
	}
}

func notifications(bot TelegramClient, client TransmissionClient, ud MessageWrapper, s settings.Settings) {
	if len(ud.Tokens()) == 0 {
		b, err := s.GetUserNotification(ud.Chat.UserName)
		if err != nil {
			send(bot, fmt.Sprintf("*notifications*: error get settings: %s", err.Error()), ud.Chat.ID)
			return
		}
		if b {
			send(bot, "*notifications* is enabled", ud.Chat.ID)
		} else {
			send(bot, "*notifications* is disabled", ud.Chat.ID)
		}
		return
	}
	switch strings.ToLower(ud.Tokens()[0]) {
	case "on", "true", "enable":
		err := s.SetUserNotification(ud.Chat.UserName, true)
		if err != nil {
			send(bot, fmt.Sprintf("*notifications*: error save settings: %s", err.Error()), ud.Chat.ID)
			return
		}
		send(bot, "*notifications*: notifications enabled", ud.Chat.ID)
	case "off", "false", "disable":
		err := s.SetUserNotification(ud.Chat.UserName, false)
		if err != nil {
			send(bot, fmt.Sprintf("*notifications*: error save settings: %s", err.Error()), ud.Chat.ID)
			return
		}
		send(bot, "*notifications*: notifications disabled", ud.Chat.ID)
	default:
		send(bot, fmt.Sprintf("*notifications*: Unknown argument `%s`", ud.CommandArguments()), ud.Chat.ID)
	}
}
