package settings_test

import (
	"github.com/zhulik/transmission-telegram/settings"
	"testing"
)

func TestSettings(t *testing.T) {
	settings, err := settings.GetSettings("/tmp/settings.db")

	if err != nil {
		t.Fail()
	}

	err = settings.SetUserID("testuser", 100500)
	if err != nil {
		t.Fail()
	}

	id, err := settings.GetUserID("testuser")
	if err != nil {
		t.Fail()
	}
	if id != 100500 {
		t.Fail()
	}
	settings.Close()
}
