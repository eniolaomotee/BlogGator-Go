package config

import "github.com/eniolaomotee/BlogGator-Go/internal/database"

type Config struct{
	DbURL string `json:"db_url"`
	UserName string `json:"current_user_name"`
}

type State struct{
	Conf *Config
	Db *database.Queries
}

type Command struct{
	Name string
	Args []string
}

type Commands struct{
	CliCommands map[string]func(*State, Command) error
}

