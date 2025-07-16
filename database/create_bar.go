package database

import (
	"fmt"
	"go-router/models"
	"go-router/utils"
	"time"

	"github.com/jackc/pgx/v5"
)

func CreateBar(conn *pgx.Conn, userInput *models.BarManual) error {
	bar := models.Bar{
		BarManual: *userInput,
	}

	node, addr, err := fetchOSM(bar.OsmID)
	if err != nil {
		return fmt.Errorf("error fetching bar from OSM: %w", err)
	}

	addrID, err := upsertLocations(conn, addr)
	if err != nil {
		return fmt.Errorf("error upserting locations: %w", err)
	}

	bar.BarOSM = addrID
	bar.BarAutoFormat = barAuto(bar.Price, bar.Size, bar.Name)

	id, err := CreateNewBar(conn, bar)
	if err != nil {
		return fmt.Errorf("error adding bar to Supabase: %w", err)
	}

	if bar.LinkedBar {
		extraDetails := ExtractBarMetadata(id, &node)
		CreateBarMetadata(conn, extraDetails)
	}

	return nil
}

func fetchOSM(osmID string) (models.NodeDetails, models.AddressParts, error) {
	// fetch location and bar details from OSM based on OSM Node
	nodeDetails, address, err := FetchBarByNode(osmID)
	if err != nil {
		return nodeDetails, address, err
	}

	return nodeDetails, address, nil
}

func upsertLocations(conn *pgx.Conn, adr models.AddressParts) (models.BarOSM, error) {
	ids := models.BarOSM{Latitude: adr.Lat, Longitude: adr.Lon}

	fylke, err := GetLocationIdByName(conn, adr.Fylke, "fylke")
	if err != nil {
		return ids, fmt.Errorf("fylke not found: %w", err)
	}

	ids.Fylke = fylke

	kommune, err := GetLocationIdByName(conn, adr.Kommune, "kommune")
	if err != nil {
		newKommune := models.Location{Name: adr.Kommune, Hierarchy: "kommune", Slug: utils.ToURL(adr.Kommune), Parent: &fylke}
		kommune, err = CreateNewLocation(conn, newKommune)
		if err != nil {
			return ids, fmt.Errorf("could not create kommune: %w", err)
		}
	}

	ids.Kommune = kommune

	if adr.Sted != "" {
		sted, err := GetLocationIdByName(conn, adr.Sted, "sted")
		if err != nil {
			newSted := models.Location{Name: adr.Sted, Hierarchy: "sted", Slug: utils.ToURL(adr.Sted), Parent: &kommune}
			sted, err = CreateNewLocation(conn, newSted)
			if err != nil {
				return ids, fmt.Errorf("could not create sted: %w", err)
			}
		}

		ids.Sted = &sted
	}

	return ids, nil
}

func barAuto(price int, size float64, name string) models.BarAutoFormat {
	auto := models.BarAutoFormat{IsActive: true, PriceUpdated: time.Now(), PriceChecked: time.Now()}
	auto.Slug = utils.ToURL(name)
	if size == 0.5 {
		auto.Pint = price
	} else {
		auto.Pint = int(float64(price) / size / 2)
	}
	return auto
}
