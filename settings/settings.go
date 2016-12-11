package settings

import (
	"github.com/boltdb/bolt"
	"strconv"
)

type Settings interface {
	SetUserID(string, int64) error
	GetUserID(string) (int64, error)
	Close()
}

const (
	users_bucket = "transmission-telegram-users"
)

type settings struct {
	db *bolt.DB
}

func (s *settings) SetUserID(username string, chatID int64) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(users_bucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(username), []byte(strconv.FormatInt(chatID, 10)))
	})
	s.db.Sync()
	return err
}

func (s *settings) GetUserID(username string) (int64, error) {
	var result int64
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(users_bucket))
		v := b.Get([]byte(username))
		var err error
		result, err = strconv.ParseInt(string(v), 10, 64)
		return err
	})
	return result, err
}

func (s *settings) Close() {
	s.db.Close()
}

func GetSettings(path string) (Settings, error) {
	db, err := bolt.Open(path, 0700, nil)
	if err != nil {
		return nil, err
	}
	return &settings{db: db}, nil
}
