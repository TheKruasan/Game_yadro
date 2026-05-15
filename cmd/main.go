package main

import (
	"bufio"
	"encoding/json"
	"os"

	"game/internal/model"
	"game/internal/report"
	"game/internal/service"
)

func main() {
	configFile, err := os.Open("../config.json")
	if err != nil {
		panic(err)
	}
	defer configFile.Close()

	var cfg model.Config

	if err := json.NewDecoder(configFile).Decode(&cfg); err != nil {
		panic(err)
	}

	eventsFile, err := os.Open("../events")
	if err != nil {
		panic(err)
	}
	defer eventsFile.Close()

	game := service.NewGame(cfg)

	scanner := bufio.NewScanner(eventsFile)

	for scanner.Scan() {
		game.ProcessLine(scanner.Text())
	}

	game.CheckResults()

	game.PrintLogs()

	report.Print(game.Players())
}
