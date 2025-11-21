package config

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"github.com/eniolaomotee/BlogGator-Go/internal/database"
	"github.com/google/uuid"
)

const configFileName = ".gatorconfig.json"

func Read()(Config, error){
	homeDir, err := os.UserHomeDir()
	if err != nil{
		return Config{}, fmt.Errorf("error getting home directory: %v", err)
	}

	filePath := filepath.Join(homeDir, configFileName)
	
	data, err := os.ReadFile(filePath)
	if err != nil{
		return Config{}, fmt.Errorf("error reading file: %v", err)
	}

	var conf Config
	
	err = json.Unmarshal(data, &conf)
	if err != nil{
		return Config{}, fmt.Errorf("couldn't unmarshall data : %v",err)
	}

	return conf,nil
}

func  (cfg *Config) SetUser (user string) error{
	cfg.UserName = user
	
	data, err := json.Marshal(cfg)
	if err != nil{
		return fmt.Errorf("error marshalling Json: %v", err)
	}
	
	homeDir, err := os.UserHomeDir()
	if err != nil{
		return fmt.Errorf("error getting home directory: %v", err)
	}
	
	path := filepath.Join(homeDir, configFileName)
	
	err = os.WriteFile(path, data, 0644)
	if err != nil{
		return fmt.Errorf("error writing to config %v",err)
	}

	return nil
}


// login function handler
func HandlerLogin(s *State, cmd Command) error {

	if len(cmd.Args) < 1{
		return fmt.Errorf("username required")
	}
	username := cmd.Args[0]

	if username == "" {
		return fmt.Errorf("username can't be empty")
	}


	_, err := s.Db.GetUser(context.Background(), username)
	if err != nil{
		return fmt.Errorf("user doesn't exist %s", err)
	}

	if err := s.Conf.SetUser(username); err != nil{
		return fmt.Errorf("error setting username")
	}

	fmt.Printf("set current user to %q\n",username)
	return  nil
}

func RegisterHandler(s *State, cmd Command) error{

	username := cmd.Args[0]

	if username == ""{
		return  fmt.Errorf("username can't be empty")
	}

	user, err := s.Db.CreateUser(context.Background(), database.CreateUserParams{
		ID:  uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name: username,
	})
	if err != nil{
		if strings.Contains(err.Error(), "duplicate key value"){
			return  fmt.Errorf("user already exists")
		}
		return fmt.Errorf("error creating user : %v", err)
	}


	if err := s.Conf.SetUser(user.Name); err != nil{
		return fmt.Errorf("error setting username")
	}

	fmt.Printf("User created successfully\n")
	fmt.Printf("User's data is %q", user)

	return nil
}

func ResetHandler(s *State, cmd Command) error{

	err := s.Db.DeleteUser(context.Background())
	if err != nil{
		return fmt.Errorf("error deleting all users : %v", err)
	}

	return nil
}

func GetAllUsersHandler(s *State, cmd Command) error{
	users, err := s.Db.GetUsers(context.Background())
	if err != nil{
		return fmt.Errorf("error getting all users: %v", err)
	}

	currentUser := s.Conf.UserName

	for _, user := range users{
		if user.Name == currentUser{
			fmt.Printf("* %s (current)\n",user.Name)
		}else{
			fmt.Printf("* %s\n", user.Name)
		}
	}


	return nil
}

// run Handler
func (c *Commands) Run (s *State, cmd Command)error{
	// commands.run: look up handler by cmd.Name; if missing, return an error; otherwise call handler(s, cmd) and return its error.

	command, ok := c.CliCommands[cmd.Name]
	if !ok{
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}else{
		return command(s, cmd)
	}
}

// Register Handler
func (c *Commands) Register (name string, f func(*State, Command)error) error{
	if name == ""{
		return fmt.Errorf("name cannot be empty")
	}

	if f == nil{
		return  fmt.Errorf("handler cannot be nil")
	}

	if c.CliCommands == nil{
		c.CliCommands = make(map[string]func(*State, Command) error)
	}

	c.CliCommands[name] = f

	return nil

}

func AggregatorService(s *State, cmd Command, user database.User) error {

	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage : agg <time_between_reqs>, e.g 'agg 1m' ")
	}

	time_between_reqs := cmd.Args[0]
	timeBetweenRequest, err := time.ParseDuration(time_between_reqs)
	if err != nil{
		return  fmt.Errorf("invalid duration :%s",err)
	}

	log.Printf("Collecting feeds every %s...", timeBetweenRequest)
	ticker := time.NewTicker(timeBetweenRequest)
	defer ticker.Stop()

	for ;; <-ticker.C{
		ScrapeFeeds(s)
	}
}

func AddFeedHandler(s *State, cmd Command, user database.User) error{

	Name := cmd.Args[0]
	UrlP := cmd.Args[1]

	feed, err := s.Db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID: uuid.New(),	
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name: Name,
		UserID: user.ID,
		Url: UrlP,

		
	})
	if err != nil{
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint"){
			return fmt.Errorf("duplicate posts")
		}
		return  fmt.Errorf("error creating feed : %s", err)
	}

	feedFollow, err := s.Db.CreateFeedFollow(context.Background(),database.CreateFeedFollowParams{
		FeedID: feed.ID,
		UserID: user.ID,
	})
	if err != nil{
		return fmt.Errorf("couldn't create feed follow: %w", err)
	}

	fmt.Printf("Feed created successfully\n")
	fmt.Printf("feed is %v\n, user is %s\n", feed,user)
	fmt.Printf("Feed followed successfully\n")
	fmt.Printf("username: %s\n, feedname: %s\n", feedFollow.UserName, feedFollow.FeedName)
	fmt.Println("=====================================")

	return nil
}


func GetAllFeeds(s *State, cmd Command) error{

	feeds, err := s.Db.GetFeeds(context.Background())
	if err != nil{
		return fmt.Errorf("error getting feeds : %s", err)
	}

	if len(feeds) == 0{
		fmt.Println("No feeds found")
		return  nil
	}

	for _, feed := range feeds{
		username, err := s.Db.GetUserFeed(context.Background(),feed.ID)
		if err != nil{
			return fmt.Errorf("error getting user feed")

		}
		fmt.Printf("name: %s url: %s user: %s\n", feed.Name, feed.Url, username)
	}
	fmt.Println("=====================================")

	return nil

}

func FollowHandler(s *State, cmd Command, user database.User) error{

	url := cmd.Args[0]
	if url == ""{
		return  fmt.Errorf("please enter a URL")
	}


	feed, err := s.Db.GetFeedByURL(context.Background(),url)
	if err != nil{
		return fmt.Errorf("couldn't get feed by URL %w", err)
	}

	feedFollow, err := s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		FeedID: feed.ID,
		UserID: user.ID,
	})
	if err != nil{
		return fmt.Errorf("couldn't create follow feed %w", err)
	}

	fmt.Printf("feedName : %s, userName following this feed : %s\n", feedFollow.FeedName, feedFollow.UserName)
	fmt.Println("=====================================")

	return nil

}

func FeedFollowingHandler(s *State, cmd Command, user database.User) error{

	feedForUser, err := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil{
		return fmt.Errorf("couldn't get feed follows for user %w",err)
	}

	if len(feedForUser) == 0 {
		fmt.Println("No feed follows found for this user")
		return nil
	}

	fmt.Printf("Feed follows for user %s:\n", user.Name)
	for _, feed := range feedForUser {
		fmt.Printf("* %s\n", feed.FeedName)
	}
	fmt.Println("=====================================")


	return  nil
}


func UnfollowHandler(s *State, cmd Command, user database.User) error {

	url := cmd.Args[0]

	urlP, err := s.Db.GetFeedByURL(context.Background(), url)
	if err != nil{
		return fmt.Errorf("couldn't get feed from URL %s", err)
	}

	err = s.Db.DeleteFeedFollowByUserAndFeed(context.Background(), database.DeleteFeedFollowByUserAndFeedParams{
		UserID: user.ID,
		FeedID: urlP.ID,
	})
	if err != nil{
		return fmt.Errorf("error unfollowing user : %s", err)
	}

	return nil
}


func ScrapeFeeds(s *State)error {

	// Get the next feed to fetch from the DB
	nextFeed, err := s.Db.GetNextFeedToFetch(context.Background())
	if err != nil{
		return fmt.Errorf("couldn't fetch next feed : %s", err)
	}

	err = s.Db.MarkFeedFetched(context.Background(), nextFeed.ID)
	if err != nil{
		return fmt.Errorf("couldn't mark feed as fetched: %s", err)
	}

	feeds, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil{
		return fmt.Errorf("couldn't fetch feed with this URL: %s", err)
	}

	for _, feed := range feeds.Channel.Item {
		PublishedAt := sql.NullTime{}
		pubDate, err := time.Parse(time.RFC1123Z, feed.PubDate); 
		if err == nil{
			PublishedAt = sql.NullTime{
				Time: pubDate,
				Valid: true,
			}
		}
		_, err = s.Db.CreatePost(context.Background(), database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Title: feed.Title,
			Description: sql.NullString{
				String: feed.Description,
				Valid: true,
			},
			PublishedAt: PublishedAt,
			Url: feed.Link,
			FeedID: nextFeed.ID,
		})
		if err != nil{
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint"){
				continue
			}
			fmt.Printf("Error creating posts :%s", err)
			continue
		}
	}

	log.Printf("Feed %s collected, %v posts found", nextFeed.Name, len(feeds.Channel.Item))
	return  nil
}



func BrowseHandler(s *State, cmd Command , user database.User)error{
	limit := 2
	if len(cmd.Args) == 1{
		if sLimit, err := strconv.Atoi(cmd.Args[0]); err != nil{
			limit = sLimit
		}else{
			return fmt.Errorf("invalid Limit: %w", err)
		}
	}
	
	posts, err := s.Db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit: int32(limit),
	})
	if err != nil{
		return fmt.Errorf("couldn't get post for user: %w", err)
	}

	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts{
		fmt.Printf("%s from %s\n", post.PublishedAt.Time.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("---- %s-----", post.Title)
		fmt.Printf("    %v\n",post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return  nil

}