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

	userErr := cfg.SetUser("Jerold")
	if userErr != nil {
		fmt.Println("There was an error setting the User", err)
		return
	}

	newCfg, err := config.Read()
	if err != nil {
		fmt.Println("There was an error reading the updated config file", err)
		return
	}

	fmt.Println(newCfg)
}
