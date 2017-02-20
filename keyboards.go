package main

import (
	"fmt"

	"gopkg.in/telegram-bot-api.v4"
)

func commandsKeyboard() *tgbotapi.ReplyKeyboardMarkup {
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

	return &commandsKeyboard
}

func torrentKeyboard(torrentID int) *tgbotapi.InlineKeyboardMarkup {
	row1 := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("stop", fmt.Sprintf("stop %d", torrentID)),
		tgbotapi.NewInlineKeyboardButtonData("start", fmt.Sprintf("start %d", torrentID)),
		tgbotapi.NewInlineKeyboardButtonData("del", fmt.Sprintf("del %d", torrentID)),
	}
	row2 := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("check", fmt.Sprintf("check %d", torrentID)),
		tgbotapi.NewInlineKeyboardButtonData("deldata", fmt.Sprintf("deldata %d", torrentID)),
	}
	commandsKeyboard := tgbotapi.NewInlineKeyboardMarkup(row1, row2)
	return &commandsKeyboard
}
