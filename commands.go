package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
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

func scrapeFeeds(s *state) error {

	nextFeedtoUpdate, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok {
			return fmt.Errorf("error getting the next Feed to fetch: %v\n", dbError.Code)
		} else {
			return fmt.Errorf("error getting the next Feed to fetch: %v\n", err)
		}
	}

	markFeedAttemptErr := s.db.MarkFeedFetched(context.Background(), nextFeedtoUpdate.ID)
	if markFeedAttemptErr != nil {
		if dbError, ok := markFeedAttemptErr.(*pq.Error); ok {
			return fmt.Errorf("there was an err marking the next feed to update: %v\n", dbError.Code)
		} else {
			return fmt.Errorf("there was an err marking the next feed to update: %v\n", markFeedAttemptErr)
		}
	}

	fetchingFeeds, err := fetchFeed(context.Background(), nextFeedtoUpdate.Url)
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok {
			return fmt.Errorf("there was an error fetching feeds: %v\n", dbError.Code)
		} else {
			return fmt.Errorf("there was an error fetching feeds: %v\n", err)
		}
	}

	for i := 0; i < len(fetchingFeeds.Channel.Item); i++ {
		fmt.Println(fetchingFeeds.Channel.Item[i].Title)

		var createPostParam database.CreatePostParams

		uuidDescription := sql.NullString{
			String: fetchingFeeds.Channel.Item[i].Description,
			Valid:  true,
		}

		timeFormats := []string{
			"Mon, 02 Jan 2006 15:04:05 GMT",
			"Mon, 02 Jan 2006 15:04:05 MST",
			"Mon, 02 Jan 2006 15:04:05 -0700",
			"Mon Jan _2 15:04:05 2006",
			"Mon Jan _2 15:04:05 MST 2006",
			"Mon Jan 02 15:04:05 -0700 2006",
			"02 Jan 06 15:04 MST",
			"02 Jan 06 15:04 -0700",
		}

		var parseTime time.Time
		for _, format := range timeFormats {
			parseTime, err = time.Parse(format, fetchingFeeds.Channel.Item[i].PubDate)
			if err == nil {
				break
			}
		}

		if err != nil {
			fmt.Printf("could not parse pub date for post: %v\n", fetchingFeeds.Channel.Item[i].Title)
			continue
		}

		uuidPublishedAt := sql.NullTime{
			Time:  parseTime,
			Valid: true,
		}

		uuidFeedID := uuid.NullUUID{
			UUID:  nextFeedtoUpdate.ID,
			Valid: true,
		}

		createPostParam.ID = uuid.New()
		createPostParam.CreatedAt = time.Now()
		createPostParam.UpdatedAt = time.Now()
		createPostParam.Title = fetchingFeeds.Channel.Item[i].Title
		createPostParam.Url = fetchingFeeds.Channel.Item[i].Link
		createPostParam.Description = uuidDescription
		createPostParam.PublishedAt = uuidPublishedAt
		createPostParam.FeedID = uuidFeedID

		createPostError := s.db.CreatePost(context.Background(), createPostParam)
		if createPostError != nil {
			if dbError, ok := createPostError.(*pq.Error); ok {
				if strings.Contains(string(dbError.Code), "23505") {
					continue
				} else {
					return fmt.Errorf("there was an error inserting post into the db: %v\n", dbError.Code)
				}
			} else {
				return fmt.Errorf("there was an error inserting post into the db: %v\n", createPostError)
			}
		}
	}

	return nil
}

func agg(s *state, cmd command) error {

	if len(cmd.args) != 1 {
		return fmt.Errorf("not enough arguments passed\n")
	}

	waitTime, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("could not convert input into time.Duration: %v\n", err)
	}

	fmt.Printf("Collecting feeds every %v\n", waitTime)

	ticker := time.NewTicker(waitTime)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			log.Printf("there was an error within scrapeFeeds: %v\n", err)
		}
	}
	return nil
}

func follow(s *state, cmd command, user database.User) error {

	uuidUser := uuid.NullUUID{
		UUID:  user.ID,
		Valid: true,
	}

	url := strings.TrimSpace(cmd.args[0])

	feedLookupResult, err := s.db.FindFeedByURL(context.Background(), url)
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok {
			return fmt.Errorf("there was an error finding the url by name: %v\n", dbError.Code)
		} else {
			return fmt.Errorf("there was an error finding the url by name: %v\n", err)
		}
	}

	var insertFeedParams database.CreateFeedFollowParams

	insertFeedParams.ID = uuid.New()
	insertFeedParams.CreatedAt = time.Now()
	insertFeedParams.UpdatedAt = time.Now()
	insertFeedParams.UserID = uuidUser.UUID
	insertFeedParams.FeedID = feedLookupResult.ID

	insertedFeed, err := s.db.CreateFeedFollow(context.Background(), insertFeedParams)
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok {
			return fmt.Errorf("there was an error inserting the new follow: %v\n", dbError.Code)
		} else {
			return fmt.Errorf("there was an error inserting the new follow: %v\n", err)
		}
	}

	fmt.Printf("%v, %v\n", insertedFeed.UserName, insertedFeed.FeedName)

	return nil
}

func browse(s *state, cmd command, user database.User) error {

	var defaultLimit int
	var limit int
	var err error
	if len(cmd.args) == 0 {
		defaultLimit = 2
	} else {
		limit, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("there was a problem parsing the provided value: %v\n", err)
		}
	}

	if limit <= 0 {
		limit = defaultLimit
	}

	uuidUserID := uuid.NullUUID{
		UUID:  user.ID,
		Valid: true,
	}

	var postParams database.GetPostsForUserParams

	postParams.UserID = uuidUserID
	postParams.Limit = int32(limit)

	posts, err := s.db.GetPostsForUser(context.Background(), postParams)
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok {
			return fmt.Errorf("there was an error getting the posts for the user: %v\n", dbError.Code)
		} else {
			return fmt.Errorf("there was an error getting the posts for the user: %v\n", err)
		}
	}

	for _, post := range posts {
		fmt.Println(post)
	}

	return nil
}

func following(s *state, cmd command, user database.User) error {

	uuidUser := uuid.NullUUID{
		UUID:  user.ID,
		Valid: true,
	}

	followingList, err := s.db.GetFeedFollowsForUser(context.Background(), uuidUser.UUID)
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok {
			return fmt.Errorf("there was an error getting the user following list: %v\n", dbError.Code)
		} else {
			return fmt.Errorf("there was an error getting the user following list: %v\n", err)
		}
	}

	for i := 0; i < len(followingList); i++ {
		fmt.Println(followingList[i])
	}

	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {

	return func(s *state, cmd command) error {
		currentUser, err := s.db.GetUser(context.Background(), s.cfg.Username)
		if err != nil {
			if dbError, ok := err.(*pq.Error); ok {
				return fmt.Errorf("there was an error getting the user: %v\n", dbError.Code)
			} else {
				return fmt.Errorf("there was an error getting the user: %v\n", err)
			}
		}
		return handler(s, cmd, currentUser)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {

	if len(cmd.args) < 2 {
		return fmt.Errorf("not enough arguments provided\n")
	}

	name := strings.TrimSpace(cmd.args[0])
	url := strings.TrimSpace(cmd.args[1])

	correctedUserID := uuid.NullUUID{
		UUID:  user.ID,
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

	var insertFeedParams database.CreateFeedFollowParams

	insertFeedParams.ID = uuid.New()
	insertFeedParams.CreatedAt = time.Now()
	insertFeedParams.UpdatedAt = time.Now()
	insertFeedParams.UserID = createFeed.UserID.UUID
	insertFeedParams.FeedID = createFeed.ID

	insertedFeed, err := s.db.CreateFeedFollow(context.Background(), insertFeedParams)
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok {
			return fmt.Errorf("there was an error inserting the new follow: %v\n", dbError.Code)
		} else {
			return fmt.Errorf("there was an error inserting the new follow: %v\n", err)
		}
	}

	fmt.Printf("%v %v", insertedFeed.UserName, insertedFeed.FeedName)

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

func handlerDelete(s *state, cmd command, user database.User) error {

	uuidUser := uuid.NullUUID{
		UUID:  user.ID,
		Valid: true,
	}

	url := cmd.args[0]

	feedID, err := s.db.FindFeedByURL(context.Background(), url)
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok {
			return fmt.Errorf("there was an error getting the url ID: %v\n", dbError.Code)
		} else {
			return fmt.Errorf("there was an error getting the url ID: %v\n", err)
		}
	}

	var deleteParams database.DeleteFollowParams

	deleteParams.UserID = uuidUser.UUID
	deleteParams.FeedID = feedID.ID

	dbErr := s.db.DeleteFollow(context.Background(), deleteParams)
	if dbErr != nil {
		if dbError, ok := dbErr.(*pq.Error); ok {
			return fmt.Errorf("there was an issue unfollowing: %v\n", dbError.Code)
		} else {
			return fmt.Errorf("there was an issue unfollowing: %v\n", dbErr)
		}
	}

	return nil
}
