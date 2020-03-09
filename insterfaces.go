package main

import (
	"github.com/pyed/transmission"
	"gopkg.in/telegram-bot-api.v4"
)

type telegramClientWrapper struct {
	bot *tgbotapi.BotAPI
}

func (bot *telegramClientWrapper) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return bot.bot.Send(c)
}
func (bot *telegramClientWrapper) GetFile(c tgbotapi.FileConfig) (tgbotapi.File, error) {
	return bot.bot.GetFile(c)
}

func (bot *telegramClientWrapper) Token() string {
	return bot.bot.Token
}

type telegramClient interface {
	Send(tgbotapi.Chattable) (tgbotapi.Message, error)
	GetFile(tgbotapi.FileConfig) (tgbotapi.File, error)
	Token() string
}

type torrentClient interface {
	GetTorrents() (transmission.Torrents, error)
	GetTorrent(int) (*transmission.Torrent, error)
	GetStats() (*transmission.Stats, error)
	AddByURL(url string) (transmission.TorrentAdded, error)
	AddByLocalFile(path string) (transmission.TorrentAdded, error)
	SetSort(transmission.Sorting)

	Version() string
	DeleteTorrent(int, bool) (string, error)
	StopAll() error
	StopTorrent(int) (string, error)
	StartAll() error
	StartTorrent(int) (string, error)
	VerifyAll() error
	VerifyTorrent(int) (string, error)
}
