package database

import (
	"fmt"
	"go-router/models"
	"go-router/utils"

	"github.com/jackc/pgx/v5"
)

func CreateBar(conn *pgx.Conn, userInput *models.BarManual) (bool, error) {
	bar := models.Bar{Name: userInput.Name}
	node, addr, err := fetchOSM(userInput.OsmID)
	if err != nil {
		return false, fmt.Errorf("Error fetching bar from OSM: ", err)
	}

	addrID, err := upsertLocations(conn, addr)
	if err != nil {
		return false, fmt.Errorf("Error upserting locations: ", err)
	}

	newBar, err := generateBarObject(node)
}

func fetchOSM(osmID string) (models.NodeDetails, models.AddressParts, error) {
	// fetch location and bar details from OSM based on OSM Node
	nodeDetails, address, err := FetchBarByNode(osmID)
	if err != nil {
		return nodeDetails, address, err
	}

	return nodeDetails, address, nil
}

func upsertLocations(conn *pgx.Conn, adr models.AddressParts) (models.AddrIDs, error) {
	ids := models.AddrIDs{}

	fylke, err := GetLocationIdByName(conn, adr.Fylke, "fylke")
	if err != nil {
		return ids, fmt.Errorf("Fylke not found: ", err)
	}

	ids.Fylke = fylke

	kommune, err := GetLocationIdByName(conn, adr.Kommune, "kommune")
	if err != nil {
		newKommune := models.Location{Name: adr.Kommune, Hierarchy: "kommune", Slug: utils.ToURL(adr.Kommune), Parent: &fylke}
		kommune, err = CreateNewLocation(conn, newKommune)
	}

	ids.Kommune = kommune

	if adr.Sted != "" {
		sted, err := GetLocationIdByName(conn, adr.Sted, "sted")
		if err != nil {
			newSted := models.Location{Name: adr.Sted, Hierarchy: "sted", Slug: utils.ToURL(adr.Sted), Parent: &kommune}
			sted, err = CreateNewLocation(conn, newSted)
		}

		ids.Sted = sted
	}

	return ids, nil
}
