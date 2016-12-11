package main

import (
	"gopkg.in/telegram-bot-api.v4"
)

func commandsKeyboard() tgbotapi.ReplyKeyboardMarkup {
	// progress stop all start all stats
	row1 := []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton("list"),
		tgbotapi.NewKeyboardButton("count"),
		tgbotapi.NewKeyboardButton("notify"),
		tgbotapi.NewKeyboardButton("speed")}

	row2 := []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton("progress"),
		tgbotapi.NewKeyboardButton("stop all"),
		tgbotapi.NewKeyboardButton("start all"),
		tgbotapi.NewKeyboardButton("stats")}
	commandsKeyboard := tgbotapi.NewReplyKeyboard(row1, row2)

	return commandsKeyboard
}
