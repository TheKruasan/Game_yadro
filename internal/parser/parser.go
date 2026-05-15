package parser

import (
	"strconv"
	"strings"
	"time"

	"game/internal/model"
)

func ParseEvent(line string) (model.Event, error) {
	parts := strings.Split(line, " ")

	rawTime := strings.Trim(parts[0], "[]")

	t, err := time.Parse("15:04:05", rawTime)
	if err != nil {
		return model.Event{}, err
	}

	playerID, err := strconv.Atoi(parts[1])
	if err != nil {
		return model.Event{}, err
	}

	eventID, err := strconv.Atoi(parts[2])
	if err != nil {
		return model.Event{}, err
	}

	extra := ""

	if len(parts) > 3 {
		extra = strings.Join(parts[3:], " ")
	}

	return model.Event{
		Time:       t,
		RawTime:    rawTime,
		PlayerID:   playerID,
		EventID:    eventID,
		ExtraParam: extra,
	}, nil
}
