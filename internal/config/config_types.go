package config

type Config struct{
	DbURL string `json:"db_url"`
	UserName string `json:"current_user_name"`
}

type State struct{
	Conf *Config
}

type Command struct{
	Name string
	Args []string
}

type Commands struct{
	CliCommands map[string]func(*State, Command) error
}