package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/NinjaCrusader/BlogAggregator/internal/config"
	"github.com/NinjaCrusader/BlogAggregator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("There was an error reading the config file on main", err)
		return
	}

	db, err := sql.Open("postgres", cfg.Url)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}

	dbQueries := database.New(db)

	newState := state{
		db:  dbQueries,
		cfg: &cfg,
	}

	commands := commands{
		commandMap: make(map[string]func(*state, command) error),
	}

	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)

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
