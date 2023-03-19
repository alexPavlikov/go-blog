package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/alexPavlikov/go-blog/api"
	"github.com/alexPavlikov/go-blog/database"
	"github.com/alexPavlikov/go-blog/models"
	"github.com/alexPavlikov/go-blog/setting"

	_ "github.com/lib/pq"
)

func init() {
	setting.Config()
}

func main() {
	fmt.Println("Listen on:", models.Cfg.ServerPort)

	db, err := database.Connect()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	api.HandleRequest()

	err = http.ListenAndServe(":"+models.Cfg.ServerPort, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}
