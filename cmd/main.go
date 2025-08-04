package main

import (
	"log"
	"net/http"
	"time"

	"go-router/database"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
)

type app struct {
	DB *pgx.Conn
}

var sessionStore = cache.New(30*time.Minute, 10*time.Minute)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	app := &app{DB: conn}
	database.InitStaticData(app.DB)

	srv := http.Server{
		Addr:    ":3000",
		Handler: app.routes(),
	}

	log.Println("Listening on :3000")
	srv.ListenAndServe()
}
