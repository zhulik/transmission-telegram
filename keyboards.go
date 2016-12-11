package main

import (
	"gopkg.in/telegram-bot-api.v4"
)

func commandsKeyboard() tgbotapi.ReplyKeyboardMarkup {
	// progress stop all start all stats
	row1 := []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton("list"),
		tgbotapi.NewKeyboardButton("speed"),
		tgbotapi.NewKeyboardButton("start all"),
		tgbotapi.NewKeyboardButton("notifications on"),
	}

	row2 := []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton("progress"),
		tgbotapi.NewKeyboardButton("stats"),
		tgbotapi.NewKeyboardButton("stop all"),
		tgbotapi.NewKeyboardButton("notifications off"),
	}
	commandsKeyboard := tgbotapi.NewReplyKeyboard(row1, row2)

	return commandsKeyboard
}
