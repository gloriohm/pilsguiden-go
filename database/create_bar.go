package database

import (
	"fmt"
	"go-router/internal/utils"
	"go-router/models"
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

	// make sure organisajson and underenhet exists, create them if not
	_, err = GetOrgID(conn, "underenheter", bar.OrgNummer)
	if err != nil {
		underenhet, err := FetchUnderenhet(bar.OrgNummer)
		if err != nil {
			return fmt.Errorf("failed fetching underenhet from brreg: %w", err)
		}
		_, err = GetOrgID(conn, "organisasjoner", underenhet.Orgnummer)
		if err != nil {
			hovedenhet, err := FetchHovedenhet(underenhet.Parent)
			if err != nil {
				return fmt.Errorf("failed fetching hovedenhet from brreg: %w", err)
			}
			if err := CreateHovedenhet(conn, hovedenhet); err != nil {
				return fmt.Errorf("failed creating hovedenhet: %w", err)
			}
		}
		if err := CreateUnderenhet(conn, underenhet); err != nil {
			return fmt.Errorf("failed creating hovedenhet: %w", err)
		}
	}

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
		newKommune := models.Location{
			BaseLocation: models.BaseLocation{
				Name: adr.Kommune,
				Slug: utils.ToURL(adr.Kommune),
			},
			Hierarchy: "kommune",
			Parent:    &fylke,
		}
		kommune, err = CreateNewLocation(conn, newKommune)
		if err != nil {
			return ids, fmt.Errorf("could not create kommune: %w", err)
		}
	}

	ids.Kommune = kommune

	if adr.Sted != "" {
		sted, err := GetLocationIdByName(conn, adr.Sted, "sted")
		if err != nil {
			newSted := models.Location{
				BaseLocation: models.BaseLocation{
					Name: adr.Sted,
					Slug: utils.ToURL(adr.Sted),
				},
				Hierarchy: "sted",
				Parent:    &kommune,
			}
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
	auto.Pint = ToPint(price, size)
	return auto
}

func ToPint(price int, size float64) int {
	var pint int
	if size == 0.5 {
		pint = price
	} else {
		pint = int(float64(price) / size / 2)
	}
	return pint
}
