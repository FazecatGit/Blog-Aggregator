package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/FazecatGit/Blog-Aggregator/internal/config"
	"github.com/FazecatGit/Blog-Aggregator/internal/database"
	"github.com/FazecatGit/Blog-Aggregator/internal/rss"
	"github.com/google/uuid"
)

type State struct {
	Config   *config.Config
	DataBase *database.Queries
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Handlers map[string]func(*State, Command) error
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.Handlers[name] = f
}

func (c *Commands) Run(s *State, cmd Command) error {
	handler, exist := c.Handlers[cmd.Name]
	if !exist {
		return fmt.Errorf("command not found: %s", cmd.Name)
	}
	return handler(s, cmd)
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: login <username>")
	}
	username := cmd.Args[0]

	user, err := s.DataBase.GetUserByName(context.Background(), username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: User %s does not exist. Please register first.", username)
		os.Exit(1)
	}

	if err := s.Config.SetUser(user.Username); err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}
	fmt.Printf("User '%s' logged in successfully.\n", user.Username)
	return nil
}

func HandlerRegister(s *State, c Command) error {
	if len(c.Args) < 1 {
		return fmt.Errorf("usage: register <username>")
	}
	username := c.Args[0]

	id := uuid.New()
	now := time.Now()

	_, err := s.DataBase.GetUserByName(context.Background(), username)
	if err == nil {
		fmt.Fprintln(os.Stderr, "user already exists")
		os.Exit(1)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("error checking user: %w", err)
	}

	user, err := s.DataBase.CreateUser(context.Background(), database.CreateUserParams{
		ID:        id,
		Username:  username,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	if err := s.Config.SetUser(user.Username); err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}

	fmt.Println("Registering user:", username)
	return nil
}

func HandlerReset(s *State, c Command) error {
	err := s.DataBase.ResetUsers(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to reset database:", err)
		os.Exit(1)
	}

	fmt.Println("Database reset successful.")
	return nil
}

func HandlerListUsers(s *State, c Command) error {
	users, err := s.DataBase.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not fetch users: %w", err)
	}

	for _, u := range users {
		if s.Config != nil && s.Config.CurrentUserName == u.Username {
			fmt.Fprintf(os.Stdout, "* %s (current)\n", u.Username)
		} else {
			fmt.Fprintf(os.Stdout, "* %s\n", u.Username)
		}
	}

	return nil
}

func HandlerAgg(s *State, c Command) error {
	feed, err := rss.FetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %w", err)
	}

	fmt.Printf("%+v\n", feed)
	return nil
}

func HandlerAddFeed(s *State, c Command, user database.User) error {
	if len(c.Args) != 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}

	name := c.Args[0]
	url := c.Args[1]

	feed, err := s.DataBase.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID, // user comes from middleware
	})
	if err != nil {
		return fmt.Errorf("error creating feed: %w", err)
	}

	_, err = s.DataBase.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to create feed follow: %w", err)
	}

	fmt.Printf("Feed created:\n")
	fmt.Printf("  ID: %s\n", feed.ID)
	fmt.Printf("  Name: %s\n", feed.Name)
	fmt.Printf("  URL: %s\n", feed.Url)
	fmt.Printf("  UserID: %s\n", feed.UserID)

	return nil
}

func HandlerListFeeds(s *State, c Command) error {
	feeds, err := s.DataBase.GetFeedsWithUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not fetch feeds: %w", err)
	}

	for _, f := range feeds {
		fmt.Printf("Name: %s\n", f.Name)
		fmt.Printf("URL: %s\n", f.Url)
		fmt.Printf("Created by: %s\n", f.UserName)
		fmt.Println("---")
	}

	return nil
}

func HandlerFollow(s *State, c Command) error {
	if len(c.Args) != 1 {
		return fmt.Errorf("usage: follow <feed_url>")
	}

	feedURL := c.Args[0]
	username := s.Config.CurrentUserName
	if username == "" {
		return fmt.Errorf("no user logged in")
	}

	user, err := s.DataBase.GetUserByName(context.Background(), username)
	if err != nil {
		return fmt.Errorf("could not fetch user: %w", err)
	}

	feed, err := s.DataBase.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("feed not found: %w", err)
	}

	feedFollow, err := s.DataBase.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("could not create feed follow: %w", err)
	}

	fmt.Printf("%s is now following %s\n", feedFollow.UserName, feedFollow.FeedName)
	return nil
}

func HandlerFollowing(s *State, c Command) error {
	username := s.Config.CurrentUserName
	if username == "" {
		return fmt.Errorf("no user logged in")
	}

	user, err := s.DataBase.GetUserByName(context.Background(), username)
	if err != nil {
		return fmt.Errorf("could not fetch user: %w", err)
	}

	follows, err := s.DataBase.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("could not fetch follow list: %w", err)
	}

	for _, f := range follows {
		fmt.Printf("* %s\n", f.FeedName)
	}

	return nil
}

func HandlerUnfollow(s *State, c Command, user database.User) error {
	if len(c.Args) != 1 {
		return fmt.Errorf("usage: unfollow <feed-url>")
	}

	url := c.Args[0]

	feed, err := s.DataBase.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("could not fetch feed: %w", err)
	}

	err = s.DataBase.DeleteFeedFollowByUserAndFeed(context.Background(), database.DeleteFeedFollowByUserAndFeedParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to unfollow feed: %w", err)
	}

	fmt.Printf("%s unfollowed feed '%s'\n", user.Username, feed.Name)
	return nil
}
