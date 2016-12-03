package main

import (
	"gopkg.in/telegram-bot-api.v4"
	"strings"
)

type UpdateWrapper tgbotapi.Update

func (w UpdateWrapper) Command() string {
	return strings.ToLower(w.Tokens()[0])
}

func (w UpdateWrapper) Tokens() []string {
	return strings.Split(w.Message.Text, " ")
}
