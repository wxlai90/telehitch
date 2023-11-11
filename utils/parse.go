package utils

import (
	"strconv"
	"strings"

	"github.com/wxlai90/telehitch/states"
)

type ParsedCallbackData struct {
	UserId    int64
	Selection states.State
}

func ParseCallbackData(data string) (ParsedCallbackData, error) {
	fields := strings.Split(data, "|")
	p := ParsedCallbackData{}

	userId, err := strconv.Atoi(fields[1])
	if err != nil {
		return p, err
	}

	p.UserId = int64(userId)
	state, err := strconv.Atoi(fields[0])
	if err != nil {
		return p, err
	}

	p.Selection = states.State(state)
	return p, nil
}
