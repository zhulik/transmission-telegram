package main

import (
	"github.com/pyed/transmission"
)

type TelegramClient interface {
}

type TransmissionClient interface {
	GetTorrents() (transmission.Torrents, error)
	GetStats() (*transmission.Stats, error)
	Version() string
	GetTorrent(int) (*transmission.Torrent, error)
	DeleteTorrent(int, bool) (string, error)
	ExecuteAddCommand(*transmission.Command) (transmission.TorrentAdded, error)
	SetSort(transmission.Sorting)
	StopAll() error
	StopTorrent(int) (string, error)
	StartAll() error
	StartTorrent(int) (string, error)
	VerifyAll() error
	VerifyTorrent(int) (string, error)
}
