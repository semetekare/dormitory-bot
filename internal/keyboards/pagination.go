package keyboards

import (
	"fmt"
	"strconv"
	"strings"
)

const DefaultPageSize = 8

type PageableButton struct {
	Text    string
	Payload string
}

func BuildPaginatedKeyboard(items []PageableButton, prefixPayload string, page, pageSize int) [][]Button {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	totalPages := (len(items) + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if page >= totalPages {
		page = totalPages - 1
	}
	if page < 0 {
		page = 0
	}

	start := page * pageSize
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}

	var buttons [][]Button
	for _, item := range items[start:end] {
		buttons = append(buttons, []Button{{Text: item.Text, Payload: item.Payload}})
	}

	if totalPages > 1 {
		var navRow []Button
		if page > 0 {
			navRow = append(navRow, Button{
				Text:    "◀",
				Payload: PagePayload(prefixPayload, page-1),
			})
		}
		navRow = append(navRow, Button{
			Text:    fmt.Sprintf("%d/%d", page+1, totalPages),
			Payload: PayloadNavBack,
		})
		if page < totalPages-1 {
			navRow = append(navRow, Button{
				Text:    "▶",
				Payload: PagePayload(prefixPayload, page+1),
			})
		}
		buttons = append(buttons, navRow)
	}

	buttons = append(buttons, []Button{{Text: "⬅️ Назад", Payload: PayloadNavBack}})
	return buttons
}

func PagePayload(prefix string, page int) string {
	return fmt.Sprintf("%s:page:%d", prefix, page)
}

func ParsePage(payload string) int {
	idx := strings.LastIndex(payload, ":page:")
	if idx == -1 {
		return 0
	}
	n, err := strconv.Atoi(payload[idx+6:])
	if err != nil {
		return 0
	}
	return n
}

func StripPageSuffix(payload string) string {
	idx := strings.LastIndex(payload, ":page:")
	if idx == -1 {
		return payload
	}
	return payload[:idx]
}

func HasPageSuffix(payload string) bool {
	return strings.LastIndex(payload, ":page:") != -1
}

func PagePrefix(payload string) string {
	base := StripPageSuffix(payload)
	page := ParsePage(payload)
	return fmt.Sprintf("%s:page:%d", base, page)
}
