package main

import (
	"fmt"

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
}
