package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/FazecatGit/Blog-Aggregator/internal/cmd"
	"github.com/FazecatGit/Blog-Aggregator/internal/config"
	"github.com/FazecatGit/Blog-Aggregator/internal/database"
	"github.com/FazecatGit/Blog-Aggregator/internal/middlewareLogin"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	dbQueries := database.New(db)

	s := &cmd.State{Config: &cfg, DataBase: dbQueries}

	commands := &cmd.Commands{
		Handlers: make(map[string]func(*cmd.State, cmd.Command) error),
	}
	commands.Register("login", cmd.HandlerLogin)
	if len(os.Args) < 2 {
		fmt.Println("Usage: <command> [arguments]")
		os.Exit(1)
	}

	commands.Register("register", cmd.HandlerRegister)
	commands.Register("reset", cmd.HandlerReset)
	commands.Register("users", cmd.HandlerListUsers)
	commands.Register("agg", cmd.HandlerAgg)
	commands.Register("feeds", cmd.HandlerListFeeds)
	commands.Register("follow", cmd.HandlerFollow)
	commands.Register("following", cmd.HandlerFollowing)
	commands.Register("addfeed", middlewareLogin.MiddlewareLoggedIn(cmd.HandlerAddFeed))
	commands.Register("unfollow", middlewareLogin.MiddlewareLoggedIn(cmd.HandlerUnfollow))

	command := cmd.Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}
	if err := commands.Run(s, command); err != nil {
		log.Fatalf("command failed: %v", err)
	}

}
