package database

import (
	"go-router/internal/stores"
	"log"

	"github.com/jackc/pgx/v5"
)

func InitStaticData(conn *pgx.Conn) {
	fylker, err := GetFylker(conn)
	if err != nil {
		log.Fatalf("failed to load fylker: %v", err)
	}

	kommuner, err := GetKommuner(conn)
	if err != nil {
		log.Fatalf("failed to load kommuner: %v", err)
	}

	steder, err := GetSteder(conn)
	if err != nil {
		log.Fatalf("failed to load steder: %v", err)
	}

	breweries, err := GetBreweries(conn)
	if err != nil {
		log.Fatalf("failed to load breweries: %v", err)
	}

	stores.AppStore.UpdateFylker(fylker)
	stores.AppStore.UpdateKommuner(kommuner)
	stores.AppStore.UpdateSteder(steder)
	stores.AppStore.UpdateBreweries(breweries)

	log.Println("✅ All static stores set")
}
