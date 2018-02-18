package main

import "strings"

const symbols = "\\*_{}[]()#+-!`"

func mdEscape(text string) string {
	for _, s := range symbols {
		ss := string(s)
		text = strings.Replace(text, ss, " ", -1)
	}
	return text
}
