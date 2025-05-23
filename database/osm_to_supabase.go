package database

import (
	"go-router/models"
	"time"
)

// should only be called if BarManual.linkedBar == true
func ExtractBarMetadata(nd *models.NodeDetails) models.BarMetadata {
	return models.BarMetadata{
		Type:         nd.Type,
		Cuisine:      nd.ExtraTags.Cuisine,
		OpeningHours: nd.ExtraTags.OpeningHours,
		Wheelchair:   nd.ExtraTags.Wheelchair,
		Website:      nd.ExtraTags.Website,
		Email:        nd.ExtraTags.Email,
		Phone:        nd.ExtraTags.Phone,
		Facebook:     nd.ExtraTags.Facebook,
		Instagram:    nd.ExtraTags.Instagram,
		LastOSMSync:  time.Now(), // Set to now or inject a param
	}
}
