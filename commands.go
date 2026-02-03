package main

import (
	"fmt"

	"github.com/NinjaCrusader/BlogAggregator/internal/config"
)

type state struct {
	config *config.Config
}

type command struct {
	name string
	args []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("no username argument provided")
	}

	s.config.Username = cmd.args[0]

	fmt.Printf("The user has been set to %v\n", s.config.Username)

	return nil
}
