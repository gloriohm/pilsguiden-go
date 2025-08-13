package database

import (
	"go-router/internal/stores"
	"go-router/models"
	"log"

	"github.com/jackc/pgx/v5"
)

func InitStaticData(conn *pgx.Conn) {
	locs, err := GetAllLocations(conn)
	if err != nil {
		log.Fatalf("failed to load locations: %v", err)
	}
	fylker, kommuner, steder := splitLocations(locs)

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

func splitLocations(in []models.Location) (A, B, C []models.Location) {
	for _, loc := range in {
		switch loc.Hierarchy {
		case "fylke":
			A = append(A, loc)
		case "kommune":
			B = append(B, loc)
		case "sted":
			C = append(C, loc)
		}
	}
	return
}
