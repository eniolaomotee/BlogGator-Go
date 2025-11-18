package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"path/filepath"
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


	_, err := s.Db.GetUserByName(context.Background(), username)
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
	if len(cmd.Args) < 1 {
		return fmt.Errorf("username required")
	}

	username := cmd.Args[0]

	if username == ""{
		return  fmt.Errorf("username can't be empty")
	}

	user, err := s.Db.CreateUser(context.Background(), database.CreateUserParams{
		ID:  uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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

func AggregatorService(s *State, cmd Command) error {

	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil{
		return fmt.Errorf("error fetching data from URL %s", err)
	}

	if feed == nil{
		return fmt.Errorf("nil feed")
	}

	fmt.Printf("Feed is %v\n", *feed)
	return nil
}

func AddFeedHandler(s *State, cmd Command) error{
	if len(cmd.Args) < 2 {
		return fmt.Errorf("url required")
	}

	Name := cmd.Args[0]
	UrlP := cmd.Args[1]

	currentUser, err := s.Db.GetUserByName(context.Background(), s.Conf.UserName)
	if err != nil{
		return fmt.Errorf("error getting user from database :%s", err)
	}

	feed, err := s.Db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID: uuid.New(),	
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: Name,
		UserID: currentUser.ID,
		Url: UrlP,

		
	})

	if err != nil{
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint"){
			return fmt.Errorf("duplicate posts")
		}
		return  fmt.Errorf("error creating feed : %s", err)
	}

	fmt.Println("feed is", feed)
	return nil
}