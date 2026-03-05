package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/NinjaCrusader/BlogAggregator/internal/config"
	"github.com/NinjaCrusader/BlogAggregator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

//state struct

type state struct {
	db  *database.Queries
	cfg *config.Config
}

//command struct

type command struct {
	name string
	args []string
}

//commands struct and helper functions

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

//commands to be used

func reset(s *state, cmd command) error {

	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		if dberror, ok := err.(*pq.Error); ok {
			return fmt.Errorf("error with delete: %v", dberror.Code)
		}
		return fmt.Errorf("error trying to delete %v", err)
	}

	fmt.Println("reset was successful")
	fmt.Println("exit status 0")

	return nil
}

func GetUsers(s *state, cmd command) error {

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok {
			return fmt.Errorf("error getting users: %v", dbError.Code)
		} else {
			return fmt.Errorf("there was an error getting users: %v", err)
		}
	}

	for i := 0; i < len(users); i++ {
		if users[i] == s.cfg.Username {
			fmt.Printf("* %v (current)\n", users[i])
		} else {
			fmt.Println(users[i])
		}
	}

	return nil
}

func agg(s *state, cmd command) error {

	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("there was an error fetching feed: %v\n", err)
	}

	fmt.Println(*feed)

	return nil
}

func follow(s *state, cmd command) error {

	return nil
}

func handlerAddFeed(s *state, cmd command) error {

	if len(cmd.args) < 2 {
		return fmt.Errorf("not enough arguments provided\n")
	}

	name := strings.TrimSpace(cmd.args[0])
	url := strings.TrimSpace(cmd.args[1])

	currentUser, err := s.db.GetUser(context.Background(), s.cfg.Username)
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok {
			return fmt.Errorf("there was an error getting the user: %v\n", dbError.Code)
		} else {
			return fmt.Errorf("there was an error getting the user: %v\n", err)
		}
	}

	correctedUserID := uuid.NullUUID{
		UUID:  currentUser.ID,
		Valid: true,
	}

	var feedParams database.AddFeedParams

	feedParams.ID = uuid.New()
	feedParams.CreatedAt = time.Now()
	feedParams.UpdatedAt = time.Now()
	feedParams.Name = name
	feedParams.Url = url
	feedParams.UserID = correctedUserID

	createFeed, err := s.db.AddFeed(context.Background(), feedParams)
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok {
			return fmt.Errorf("there was an error adding feed to the db: %v\n", dbError.Code)
		} else {
			return fmt.Errorf("there was an error adding feed to the db: %v\n", err)
		}
	}

	fmt.Println(createFeed.ID)
	fmt.Println(createFeed.CreatedAt)
	fmt.Println(createFeed.UpdatedAt)
	fmt.Println(createFeed.Name)
	fmt.Println(createFeed.Url)
	fmt.Println(createFeed.UserID)

	return nil
}

func handlerPrintFeeds(s *state, cmd command) error {

	feeds, err := s.db.PrintFeeds(context.Background())
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok {
			return fmt.Errorf("there was an error printing feeds from the db: %v\n", dbError.Code)
		} else {
			return fmt.Errorf("there was an error printing feeds from the db: %v\n", err)
		}
	}

	for i := 0; i < len(feeds); i++ {
		fmt.Println(feeds[i])
	}

	return nil
}

func handlerRegister(s *state, cmd command) error {

	if len(cmd.args) < 1 {
		return fmt.Errorf("no argument provided")
	}

	var userParams database.CreateUserParams

	userParams.ID = uuid.New()
	userParams.CreatedAt = time.Now()
	userParams.UpdatedAt = time.Now()
	userParams.Name = cmd.args[0]

	createdUser, err := s.db.CreateUser(context.Background(), userParams)
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok {
			if dbError.Code == "23505" {
				return fmt.Errorf("a user with this name already exists: %v\n", dbError)
			}
			return fmt.Errorf("error creating user: %v\n", dbError.Code)
		} else {
			return fmt.Errorf("error creating user: %v\n", err)
		}
	}

	s.cfg.SetUser(createdUser.Name)

	fmt.Printf("The user %v was created %v\n", createdUser.Name, createdUser)

	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: login <username>")
	}

	username := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("this user does not exist\n")
		}

		if dbError, ok := err.(*pq.Error); ok {
			return fmt.Errorf("error with query: %v\n", dbError.Code)
		}
		return fmt.Errorf("error with query: %v\n", err)
	}

	if err := s.cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Printf("The user has been set to %v\n", s.cfg.Username)

	return nil
}
