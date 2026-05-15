package model

import "time"

type Config struct {
	Floors   int    `json:"Floors"`
	Monsters int    `json:"Monsters"`
	OpenAt   string `json:"OpenAt"`
	Duration int    `json:"Duration"`
}

type Event struct {
	Time       time.Time
	RawTime    string
	PlayerID   int
	EventID    int
	ExtraParam string
}

type State string

const (
	Success State = "SUCCESS"
	Fail    State = "FAIL"
	Disqual State = "DISQUAL"
)

type Player struct {
	ID int

	Registered bool
	InDungeon  bool
	Dead       bool

	InBossRoom bool

	State State

	HP int

	CurrentFloor int

	FloorKills    map[int]int
	ClearedFloors map[int]bool

	FloorsTimes map[int]time.Duration

	EnterTime *time.Time
	LeaveTime *time.Time

	BossKilled bool

	BossKillTime   time.Duration
	StartBossFight *time.Time

	EnterCurrentFloorTime *time.Time
}
