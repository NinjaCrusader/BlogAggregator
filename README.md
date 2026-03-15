# BlogAggregator

BlogAggregator is a CLI-based RSS feed aggregator written in Go. [file:3][file:2]  
It lets you register users, add and follow RSS feeds, periodically scrape new posts into a PostgreSQL database, and browse posts for a given user.

## Features

- User registration and login stored in Postgres via sqlc-generated queries.
- Add RSS feeds and automatically follow them as the current user.  
- Periodic aggregation of RSS feeds with robust RSS parsing and multiple date formats. 
- Follow and unfollow feeds, and list what the current user is following. 
- Browse recent posts for the current user with an adjustable limit.  
- Simple configuration system for database URL and current username.

## Project Structure

- `main.go` – Application entrypoint; loads config, connects to Postgres, wires command handlers, and dispatches CLI commands. [file:3][file:2]  
- `commands.go` – Core command handling, state management, and all CLI handlers (`register`, `login`, `addfeed`, `agg`, `follow`, `unfollow`, `following`, `users`, `browse`, `reset`, etc.).
- `rss_feed.go` – RSS parsing helpers, including HTTP fetching, XML unmarshalling, HTML unescaping, and pubDate normalization.   
- `sqlc.yaml` – sqlc configuration pointing at the schema and query folders and generating Go code into `internal/database` for a PostgreSQL backend.
- `internal/config` – Configuration package (referenced but not included here) used for reading DB URL and tracking the current username.  
- `internal/database` – sqlc-generated query layer (types and methods such as `CreateUser`, `GetUsers`, `AddFeed`, `CreateFeedFollow`, `GetPostsForUser`, etc.).

## Prerequisites

- Go (version compatible with modules in this project).
- PostgreSQL database instance.  
- A valid schema and queries under `sql/schema` and `sql/queries` consumable by sqlc. 
- sqlc installed if you need to regenerate the `internal/database` package. 

You are expected to have a PostgreSQL DSN (URL) that the config layer can read (for example from a config file or environment variable, depending on how `internal/config` is implemented). 
