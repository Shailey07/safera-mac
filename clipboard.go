package main

import (
	"errors"
	"github.com/atotto/clipboard"
)

func readClipboard() (string, error) {
	text, err := clipboard.ReadAll()
	if err != nil {
		return "", errors.New("clipboard read failed: " + err.Error())
	}
	return text, nil
}