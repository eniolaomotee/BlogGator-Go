package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

	fmt.Printf("set current user to %q\n", username)
	fmt.Printf("User's data is %q", user)

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