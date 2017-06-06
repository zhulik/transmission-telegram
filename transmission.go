package main

import "github.com/pyed/transmission"

type transmissionClient struct {
	client *transmission.TransmissionClient
}

func (client transmissionClient) DeleteTorrent(id int, wd bool) (string, error) {
	return client.client.DeleteTorrent(id, wd)
}

func (client transmissionClient) AddByURL(url string) (transmission.TorrentAdded, error) {
	cmd := transmission.NewAddCmdByURL(url)
	return client.client.ExecuteAddCommand(cmd)
}

func (client transmissionClient) GetStats() (*transmission.Stats, error) {
	return client.client.GetStats()
}

func (client transmissionClient) GetTorrent(id int) (*transmission.Torrent, error) {
	return client.client.GetTorrent(id)
}

func (client transmissionClient) GetTorrents() (transmission.Torrents, error) {
	return client.client.GetTorrents()
}

func (client transmissionClient) SetSort(s transmission.Sorting) {
	client.client.SetSort(s)
}

func (client transmissionClient) StopAll() error {
	return client.client.StopAll()
}

func (client transmissionClient) StopTorrent(id int) (string, error) {
	return client.client.StopTorrent(id)
}

func (client transmissionClient) StartAll() error {
	return client.client.StartAll()
}

func (client transmissionClient) StartTorrent(id int) (string, error) {
	return client.client.StartTorrent(id)
}

func (client transmissionClient) VerifyAll() error {
	return client.client.VerifyAll()
}

func (client transmissionClient) VerifyTorrent(id int) (string, error) {
	return client.client.VerifyTorrent(id)
}

func (client transmissionClient) Version() string {
	return client.client.Version()
}
