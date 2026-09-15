package types

import "github.com/dormitory-bot/internal/keyboards"

type MessageResponse struct {
	Text    string
	Buttons [][]keyboards.Button
}

func TextResponse(text string) *MessageResponse {
	return &MessageResponse{Text: text, Buttons: nil}
}

func ResponseWithButtons(text string, rows [][]keyboards.Button) *MessageResponse {
	return &MessageResponse{Text: text, Buttons: rows}
}
