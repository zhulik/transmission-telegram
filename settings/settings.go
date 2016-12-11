package settings

import (
	"github.com/boltdb/bolt"
	// "log"
	"strconv"
)

const (
	users_bucket  = "transmission-telegram-users"
	notify_bucket = "transmission-telegram-notify"
)

type Settings interface {
	SetUserID(string, int64) error
	GetUserID(string) (int64, error)
	SetUserNotification(string, bool) error
	GetUserNotification(string) (bool, error)
	Close()
}

func GetSettings(path string) (Settings, error) {
	db, err := bolt.Open(path, 0700, nil)
	if err != nil {
		return nil, err
	}
	return &settings{db: db}, nil
}

type settings struct {
	db *bolt.DB
}

func (s *settings) SetUserID(username string, chatID int64) error {
	err := s.set(users_bucket, username, strconv.FormatInt(chatID, 10))
	return err
}

func (s *settings) GetUserID(username string) (int64, error) {
	val, err := s.get(users_bucket, username)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(string(val), 10, 64)
}

func (s *settings) SetUserNotification(username string, notify bool) error {
	return s.set(notify_bucket, username, strconv.FormatBool(notify))
}

func (s *settings) GetUserNotification(username string) (bool, error) {
	v, err := s.get(notify_bucket, username)
	if err != nil {
		return false, err
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, nil
	}
	return b, nil
}

func (s *settings) set(bucket string, key string, value string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), []byte(value))
	})
	s.db.Sync()
	return err
}

func (s *settings) get(bucket string, key string) (string, error) {
	var result string
	err := s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		result = string(b.Get([]byte(key)))
		return nil
	})
	return result, err
}

func (s *settings) Close() {
	s.db.Close()
}
