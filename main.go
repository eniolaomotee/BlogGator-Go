package main

import (
	"fmt"
	"log"
	"os"

	"github.com/eniolaomotee/BlogGator-Go/internal/config"
)



func main(){
	// Read Config
	cfg, err := config.Read()
	if err != nil{
		log.Fatalf("error reading config :%v", err)
	}

	fmt.Println("cfg", cfg)

	// Build State
	state := &config.State{
		Conf: &cfg,
	}

	fmt.Println("state", state)

	// Build commands registry
	cmds := &config.Commands{}
	if err := cmds.Register("login", config.HandlerLogin); err != nil{
		log.Fatalf("error register config :%v", err)
	}


	// parse Args
	if len(os.Args) < 2 {
		return 
	}

	name := os.Args[1]
	args := os.Args[2:]
	cmd := config.Command{Name: name, Args: args}


	// run 
	if err := cmds.Run(state,cmd); err != nil{
		return 
	}

	os.Exit(1)

}
