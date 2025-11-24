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

type RSSFeed struct{
	Channel struct{
		Title string `xml:"title"`
		Link string `xml:"link"`
		Description string `xml:"description"`
		Item []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct{
	Title string `xml:"title"`
	Link string `xml:"link"`
	Description string `xml:"description"`
	PubDate string `xml:"pubDate"`
}

type BrowseFlags struct{
	Limit int
	SortBy     string
	Order      string
	FeedFilter string
}