package report

import (
	"fmt"
	"time"

	"game/internal/model"
)

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	return fmt.Sprintf(
		"%02d:%02d:%02d",
		hours,
		minutes,
		seconds,
	)
}

func Print(players map[int]*model.Player) {
	fmt.Println()
	fmt.Println("Final report:")

	for id := 1; id <= len(players); id++ {

		p := players[id]

		totalTime := time.Duration(0)

		if p.EnterTime != nil &&
			p.LeaveTime != nil {

			totalTime =
				p.LeaveTime.Sub(*p.EnterTime)
		}

		totalFloorTime := time.Duration(0)
		floors := 0

		for _, duration := range p.FloorsTimes {
			totalFloorTime += duration
			floors++
		}

		avgFloor := time.Duration(0)

		if floors > 0 {
			avgFloor =
				totalFloorTime /
					time.Duration(floors)
		}

		bossTime := time.Duration(0)

		if p.BossKilled {
			bossTime = p.BossKillTime
		}

		fmt.Printf(
			"[%s] %d [%s, %s, %s] HP:%d\n",
			p.State,
			p.ID,
			formatDuration(totalTime),
			formatDuration(avgFloor),
			formatDuration(bossTime),
			p.HP,
		)
	}
}
