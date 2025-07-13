package database

import (
	"go-router/internal/stores"
	"go-router/models"
)

func InitiateCreateBar(sessID string, userInput *models.BarManual) (models.AddressParts, error) {
	// fetch location and bar details from OSM based on OSM Node
	nodeDetails, address, err := FetchBarByNode(userInput.OsmID)
	if err != nil {
		return address, err
	}

	stores.SetBarStore(sessID, *userInput)
	stores.SetAddressStore(sessID, address)

	if userInput.LinkedBar {
		extraDetails := ExtractBarMetadata(&nodeDetails)
		stores.SetMetaStore(sessID, extraDetails)
	}

	return address, nil
}
