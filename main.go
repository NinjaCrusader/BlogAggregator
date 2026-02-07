package main

import (
	"fmt"
	"log"
	"os"

	"github.com/NinjaCrusader/BlogAggregator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("There was an error reading the config file on main", err)
		return
	}

	newState := state{
		config: &cfg,
	}

	commands := commands{
		commandMap: make(map[string]func(*state, command) error),
	}

	commands.register("login", handlerLogin)

	if len(os.Args) < 2 {
		log.Fatalf("not enough arguments provided")
	}

	command := command{
		name: os.Args[1],
		args: os.Args[2:],
	}

	if err := commands.run(&newState, command); err != nil {
		log.Fatal(err)
	}

}
