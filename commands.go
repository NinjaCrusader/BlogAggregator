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

type commands struct {
	commandMap map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	if commandToRun, ok := c.commandMap[cmd.name]; ok {
		err := commandToRun(s, cmd)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("command doesn't exist")
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandMap[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: login <username>")
	}

	s.config.Username = cmd.args[0]

	if err := s.config.SetUser(cmd.args[0]); err != nil {
		return err
	}

	fmt.Printf("The user has been set to %v\n", s.config.Username)

	return nil
}
