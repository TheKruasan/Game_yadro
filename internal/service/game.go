package service

import (
	"fmt"
	"strconv"
	"time"

	"game/internal/model"
	"game/internal/parser"
)

type Game struct {
	cfg model.Config

	players map[int]*model.Player
	logs    []string

	openTime  time.Time
	closeTime time.Time
}

func NewGame(cfg model.Config) *Game {
	openTime, _ := time.Parse("15:04:05", cfg.OpenAt)

	closeTime := openTime.Add(
		time.Duration(cfg.Duration) * time.Hour,
	)

	return &Game{
		cfg:       cfg,
		players:   make(map[int]*model.Player),
		openTime:  openTime,
		closeTime: closeTime,
	}
}

func (g *Game) Players() map[int]*model.Player {
	return g.players
}

func (g *Game) CheckResults() {
	for _, player := range g.players {

		if player.State != model.Disqual || !player.Dead {

			if g.allFloorsCleared(player) {
				player.State = model.Success
			}
		}
	}
}

func (g *Game) ProcessLine(line string) {
	event, err := parser.ParseEvent(line)
	if err != nil {
		return
	}

	player := g.getOrCreatePlayer(event.PlayerID)

	//если умер то мы заканчиваем испытание и не можем польше ничего писать
	if player.Dead {
		return
	}

	//если дисквалифицирован то мы заканчиваем испытание и не можем польше ничего писать
	if player.State == model.Disqual {
		return
	}

	//если вышел то мы заканчиваем испытание и не можем польше ничего писать
	if player.LeaveTime != nil {
		return
	}

	//если не регистрация то мы проверям на существование регистрации у игрока
	//и проверяем время ивента
	if event.EventID != 1 {

		if !player.Registered {
			player.State = model.Disqual

			g.log(
				event.RawTime,
				fmt.Sprintf(
					"Player [%d] is disqualified",
					player.ID,
				),
			)
			return
		}

		if event.Time.Before(g.openTime) {

			g.impossibleMove(player, event)
			return
		}

		if event.Time.After(g.closeTime) {

			g.impossibleMove(player, event)
			return
		}
	}
	switch event.EventID {

	case 1:
		g.register(player, event)

	case 2:
		g.enterDungeon(player, event)

	case 3:
		g.killMonster(player, event)

	case 4:
		g.nextFloor(player, event)

	case 5:
		g.prevFloor(player, event)

	case 6:
		g.enterBoss(player, event)

	case 7:
		g.killBoss(player, event)

	case 8:
		g.leaveDungeon(player, event)

	case 9:
		g.cannotContinue(player, event)

	case 10:
		g.restoreHealth(player, event)

	case 11:
		g.takeDamage(player, event)
	}
}

func (g *Game) getOrCreatePlayer(id int) *model.Player {
	if p, ok := g.players[id]; ok {
		return p
	}

	player := &model.Player{
		ID:    id,
		HP:    100,
		State: model.Fail,

		FloorKills:    make(map[int]int),
		ClearedFloors: make(map[int]bool),
		FloorsTimes:   make(map[int]time.Duration),
		// FloorClearTimes: make(map[int]time.Duration),
	}

	g.players[id] = player

	return player
}

func (g *Game) log(timeStr string, text string) {
	g.logs = append(
		g.logs,
		fmt.Sprintf("[%s] %s", timeStr, text),
	)
}

func (g *Game) PrintLogs() {
	for _, log := range g.logs {
		fmt.Println(log)
	}
}

func (g *Game) register(
	p *model.Player,
	e model.Event,
) {
	if p.Registered {

		g.impossibleMove(p, e)
		return
	}

	p.Registered = true

	g.log(
		e.RawTime,
		fmt.Sprintf(
			"Player [%d] registered",
			p.ID,
		),
	)
}

func (g *Game) enterDungeon(
	p *model.Player,
	e model.Event,
) {
	if p.InDungeon || p.LeaveTime != nil {
		g.impossibleMove(p, e)
		return
	}

	p.InDungeon = true
	p.CurrentFloor = 1

	enterTime := e.Time
	p.EnterTime = &enterTime

	p.FloorsTimes[1] = time.Duration(0)
	p.EnterCurrentFloorTime = &e.Time

	g.log(
		e.RawTime,
		fmt.Sprintf(
			"Player [%d] entered the dungeon",
			p.ID,
		),
	)
}

func (g *Game) killMonster(
	p *model.Player,
	e model.Event,
) {

	if p.ClearedFloors[p.CurrentFloor] || !p.InDungeon || p.CurrentFloor == g.cfg.Floors {
		g.impossibleMove(p, e)
		return
	}

	p.FloorKills[p.CurrentFloor]++

	g.log(
		e.RawTime,
		fmt.Sprintf(
			"Player [%d] killed the monster",
			p.ID,
		),
	)

	if p.FloorKills[p.CurrentFloor] ==
		g.cfg.Monsters {

		p.ClearedFloors[p.CurrentFloor] = true
		duration := p.FloorsTimes[p.CurrentFloor] + e.Time.Sub(*p.EnterCurrentFloorTime)
		p.FloorsTimes[p.CurrentFloor] = duration
	}

}

func (g *Game) nextFloor(
	p *model.Player,
	e model.Event,
) {

	if p.CurrentFloor == g.cfg.Floors || !p.InDungeon {

		g.impossibleMove(p, e)
		return
	}

	if !p.ClearedFloors[p.CurrentFloor] {
		p.FloorsTimes[p.CurrentFloor] += e.Time.Sub(*p.EnterCurrentFloorTime)
	}
	p.CurrentFloor++

	p.EnterCurrentFloorTime = &e.Time

	g.log(
		e.RawTime,
		fmt.Sprintf(
			"Player [%d] went to the next floor",
			p.ID,
		),
	)
}

func (g *Game) prevFloor(
	p *model.Player,
	e model.Event,
) {

	if p.CurrentFloor == 1 || !p.InDungeon {

		g.impossibleMove(p, e)
		return
	}

	if p.InBossRoom && !p.BossKilled {
		p.InBossRoom = false
		p.BossKillTime += e.Time.Sub(*p.StartBossFight)
		p.StartBossFight = nil
	} else if !p.ClearedFloors[p.CurrentFloor] {
		p.FloorsTimes[p.CurrentFloor] += e.Time.Sub(*p.EnterCurrentFloorTime)
	}

	p.CurrentFloor--

	p.EnterCurrentFloorTime = &e.Time

	g.log(
		e.RawTime,
		fmt.Sprintf(
			"Player [%d] went to the previous floor",
			p.ID,
		),
	)
}

func (g *Game) enterBoss(
	p *model.Player,
	e model.Event,
) {
	if !p.InDungeon || p.CurrentFloor != g.cfg.Floors || p.InBossRoom {

		g.impossibleMove(p, e)
		return
	}

	p.InBossRoom = true
	bossTime := e.Time
	p.StartBossFight = &bossTime

	g.log(
		e.RawTime,
		fmt.Sprintf(
			"Player [%d] entered the boss's floor",
			p.ID,
		),
	)
}
func (g *Game) allFloorsCleared(
	p *model.Player,
) bool {

	for floor := 1; floor <= g.cfg.Floors; floor++ {

		if !p.ClearedFloors[floor] {
			return false
		}
	}

	return true
}
func (g *Game) killBoss(
	p *model.Player,
	e model.Event,
) {

	if !p.InBossRoom || p.BossKilled || !p.InDungeon {

		g.impossibleMove(p, e)
		return
	}

	p.BossKilled = true
	p.BossKillTime += e.Time.Sub(*p.StartBossFight)
	p.ClearedFloors[p.CurrentFloor] = true

	g.log(
		e.RawTime,
		fmt.Sprintf(
			"Player [%d] killed the boss",
			p.ID,
		),
	)
}

func (g *Game) leaveDungeon(
	p *model.Player,
	e model.Event,
) {
	if !p.InDungeon {

		g.impossibleMove(p, e)
		return
	}

	leaveTime := e.Time
	p.LeaveTime = &leaveTime
	p.InDungeon = false

	g.log(
		e.RawTime,
		fmt.Sprintf(
			"Player [%d] left the dungeon",
			p.ID,
		),
	)
}

func (g *Game) cannotContinue(
	p *model.Player,
	e model.Event,
) {
	p.State = model.Disqual

	g.log(
		e.RawTime,
		fmt.Sprintf(
			"Player [%d] cannot continue due to [%s]",
			p.ID,
			e.ExtraParam,
		),
	)
}

func (g *Game) restoreHealth(
	p *model.Player,
	e model.Event,
) {
	if p.Dead {

		g.impossibleMove(p, e)
		return
	}

	hp, _ := strconv.Atoi(e.ExtraParam)

	p.HP += hp

	if p.HP > 100 {
		p.HP = 100
	}

	g.log(
		e.RawTime,
		fmt.Sprintf(
			"Player [%d] has restored [%d] of health",
			p.ID,
			hp,
		),
	)
}

func (g *Game) takeDamage(
	p *model.Player,
	e model.Event,
) {
	// если умер, этаж зачищен или игрок не в данже то не можем получить урон (а тае же если мы на этаже босса но не зашли к нему)
	if p.Dead || p.ClearedFloors[p.CurrentFloor] || !p.InDungeon || (!p.InBossRoom && p.CurrentFloor == g.cfg.Floors) {

		g.impossibleMove(p, e)
		return
	}

	damage, _ := strconv.Atoi(e.ExtraParam)

	p.HP -= damage

	g.log(
		e.RawTime,
		fmt.Sprintf(
			"Player [%d] recieved [%d] of damage",
			p.ID,
			damage,
		),
	)

	if p.HP <= 0 {

		p.HP = 0
		p.Dead = true
		p.State = model.Fail

		g.log(
			e.RawTime,
			fmt.Sprintf(
				"Player [%d] is dead",
				p.ID,
			),
		)
	}
}

func (g *Game) impossibleMove(
	p *model.Player,
	e model.Event,
) {
	g.log(
		e.RawTime,
		fmt.Sprintf(
			"Player [%d] makes imposible move [%d]",
			p.ID,
			e.EventID,
		),
	)
}
