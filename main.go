package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"github.com/eniolaomotee/BlogGator-Go/internal/config"
	"github.com/eniolaomotee/BlogGator-Go/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)



func main(){
	//Load dotenv
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	// Database connections
	db, err := sql.Open("postgres",dbUrl)
	if err != nil{
		log.Fatalf("error connecting to database %v", err)
	}

	defer db.Close()

	dbQueries := database.New(db)


	// Read config
	cfg, err := config.Read()
	if err != nil{
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	//Build State
	state := &config.State{
		Db : dbQueries,
		Conf: &cfg,
		
	}
	// Build Command Registry
	cmds := &config.Commands{}

	cmds.Register("login", config.HandlerLogin)
	cmds.Register("register", config.ArgumentValidationMiddleware(config.RegisterHandler, 1))
	cmds.Register("reset", config.ResetHandler)
	cmds.Register("users", config.GetAllUsersHandler)
	cmds.Register("agg", config.AggregatorService)
	cmds.Register("feeds", config.GetAllFeeds)
	cmds.Register("follow", config.ArgumentValidationMiddleware(config.MiddlewareLoggedIn(config.FollowHandler),1))
	cmds.Register("addfeed", config.ArgumentValidationMiddleware(config.MiddlewareLoggedIn(config.AddFeedHandler),2))
	cmds.Register("following", config.MiddlewareLoggedIn(config.FeedFollowingHandler))


	// Parse Args
	if len(os.Args) < 2{
		log.Fatalf("Usage: cli <command> [args...]")
	}

	name :=  os.Args[1]
	args := os.Args[2:]

	cmd := config.Command{Name: name, Args: args}
	// Run
	if err := cmds.Run(state,cmd); err != nil{
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}


}
