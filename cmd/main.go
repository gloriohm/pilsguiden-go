package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go-router/internal/platform/database"

	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
)

var sessionStore = cache.New(30*time.Minute, 10*time.Minute)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	dsn := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=require",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PW"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	ctx := context.Background()
	pool, err := database.NewPool(ctx, dsn)

	if err != nil {
		log.Fatal(err)
	}

	appRepo := app.NewAppRepo(pool)

	app := &app{Pool: pool}
	database.InitStaticData(app.Pool)

	srv := http.Server{
		Addr:    ":3000",
		Handler: app.routes(),
	}

	log.Println("Listening on :3000")
	srv.ListenAndServe()
}
