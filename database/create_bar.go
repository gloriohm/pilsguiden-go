package database

import (
	"go-router/internal/stores"
	"go-router/models"
)

func InitiateCreateBar(userInput *models.BarManual) (*models.AddressParts, error) {
	store := &stores.CreateBarStore{}
	bar := models.Bar{BarManual: *userInput}

	// fetch location and bar details from OSM based on OSM Node
	nodeDetails, address, err := FetchBarByNode(bar.OsmID)
	if err != nil {
		return nil, err
	}

	store.SetBar(&bar)
	store.SetAddress(address)

	if userInput.LinkedBar {
		extraDetails := ExtractBarMetadata(nodeDetails)
		store.SetMetadata(&extraDetails)
	}

	return address, nil
}
