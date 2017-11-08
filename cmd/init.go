package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/m0cchi/gfalcon/model"

	"os"
)

const ServiceName = "sahohime"
const PostActionID = "postable"

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

	service, err := model.CreateService(db, ServiceName)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	_, err = model.CreateAction(db, service.IID, PostActionID)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
}
