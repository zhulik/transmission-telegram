package settings_test

import (
	"github.com/zhulik/transmission-telegram/settings"
	"os"
	"testing"
)

const (
	path = "/tmp/settings.db"
)

func TestSetGetUserID(t *testing.T) {
	os.Remove(path)
	settings, err := settings.GetSettings(path)

	if err != nil {
		t.Fatal(err)
	}

	err = settings.SetUserID("testuser", 100500)
	if err != nil {
		t.Fatal(err)
	}

	id, err := settings.GetUserID("testuser")
	if err != nil {
		t.Fatal(err)
	}
	if id != 100500 {
		t.Fatal("Wrong id returned")
	}
	settings.Close()
}

func TestGetUserID(t *testing.T) {
	os.Remove(path)
	settings, err := settings.GetSettings(path)

	if err != nil {
		t.Fatal(err)
	}

	_, err = settings.GetUserID("testuser")
	if err == nil {
		t.Fatal(err)
	}
	settings.Close()
}

func TestSetGetNotification(t *testing.T) {
	os.Remove(path)
	settings, err := settings.GetSettings(path)

	if err != nil {
		t.Fail()
	}

	err = settings.SetUserNotification("testuser", true)
	if err != nil {
		t.Fatal(err)
	}

	n, err := settings.GetUserNotification("testuser")
	if err != nil {
		t.Fatal(err)
	}
	if !n {
		t.Fatal("Wrong value returned")
	}
	settings.Close()
}

func TestGetNotification(t *testing.T) {
	os.Remove(path)
	settings, err := settings.GetSettings(path)

	if err != nil {
		t.Fail()
	}

	n, err := settings.GetUserNotification("testuser")
	if err != nil {
		t.Fatal(err)
	}
	if n {
		t.Fatal("Wrong value returned")
	}
	settings.Close()
}
