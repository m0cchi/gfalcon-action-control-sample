package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/m0cchi/gfalcon/model"

	"os"
)

const ServiceName = "sahohime"
const PostActionID = "postable"
const TeamID = "gfalcon"
const UserID = "sahohime"

func main() {
	datasource := os.Getenv("DATASOURCE")
	if datasource == "" {
		fmt.Println("not specify datasource")
		os.Exit(1)
	}
	db, err := sqlx.Connect("mysql", datasource)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	service, err := model.GetService(db, ServiceName)
	if err != nil {
		fmt.Printf("failed get service: %v\n", err)
		os.Exit(1)
	}

	action, err := model.GetAction(db, service.IID, PostActionID)
	if err != nil {
		fmt.Printf("faield get action: %v\n", err)
		os.Exit(1)
	}

	team, err := model.GetTeam(db, TeamID)
	if err != nil {
		fmt.Printf("failed get team: %v\n", err)
		os.Exit(1)
	}

	user, err := model.CreateUser(db, team.IID, UserID)
	if err != nil {
		fmt.Printf("failed create User: %v\n", err)
		os.Exit(1)
	}

	password := "secret"
	err = user.UpdatePassword(db, password)
	if err != nil {
		fmt.Printf("failed create password: %v\n", err)
		os.Exit(1)
	}

	err = model.CreateActionLink(db, action, user)
	if err != nil {
		fmt.Printf("failed create ActionLink: %v\n", err)
		os.Exit(1)
	}
}
