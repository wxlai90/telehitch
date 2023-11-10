package utils

import (
	"strconv"
	"strings"
)

type ParsedCallbackData struct {
	UserId    int64
	Selection string
}

func ParseCallbackData(data string) (ParsedCallbackData, error) {
	fields := strings.Split(data, "|")
	p := ParsedCallbackData{}

	userId, err := strconv.Atoi(fields[1])
	if err != nil {
		return p, err
	}

	p.UserId = int64(userId)
	p.Selection = fields[0]

	return p, nil
}
