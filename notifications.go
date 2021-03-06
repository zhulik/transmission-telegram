package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pyed/transmission"
	"github.com/zhulik/transmission-telegram/settings"
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

func sendFinishedTorrent(bot telegramClient, t *transmission.Torrent, chatID int64) {
	msg := fmt.Sprintf("*%d* `%s` is finished!", t.ID, ellipsisString(t.Name, 25))
	send(bot, msg, chatID, true)
	log.Println("Finished torrent was sent")
}

func notifyFinished(bot telegramClient, client torrentClient, masters []string, s settings.Settings) {
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

func notifications(bot telegramClient, client torrentClient, ud messageWrapper, s settings.Settings) {
	if len(ud.Tokens()) == 0 {
		b, err := s.GetUserNotification(ud.Chat.UserName)
		if err != nil {
			send(bot, fmt.Sprintf("*notifications*: error get settings: %s", err.Error()), ud.Chat.ID, true)
			return
		}
		if b {
			send(bot, "*notifications* is enabled", ud.Chat.ID, true)
		} else {
			send(bot, "*notifications* is disabled", ud.Chat.ID, true)
		}
		return
	}
	switch strings.ToLower(ud.Tokens()[0]) {
	case "on", "true", "enable":
		err := s.SetUserNotification(ud.Chat.UserName, true)
		if err != nil {
			send(bot, fmt.Sprintf("*notifications*: error save settings: %s", err.Error()), ud.Chat.ID, true)
			return
		}
		send(bot, "*notifications*: notifications enabled", ud.Chat.ID, true)
	case "off", "false", "disable":
		err := s.SetUserNotification(ud.Chat.UserName, false)
		if err != nil {
			send(bot, fmt.Sprintf("*notifications*: error save settings: %s", err.Error()), ud.Chat.ID, true)
			return
		}
		send(bot, "*notifications*: notifications disabled", ud.Chat.ID, true)
	default:
		send(bot, fmt.Sprintf("*notifications*: Unknown argument `%s`", ud.CommandArguments()), ud.Chat.ID, true)
	}
}
