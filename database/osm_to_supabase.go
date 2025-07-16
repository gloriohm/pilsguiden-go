package database

import (
	"go-router/models"
	"time"
)

func ExtractBarMetadata(id int, nd *models.NodeDetails) models.BarMetadata {
	return models.BarMetadata{
		BarID:        id,
		Type:         nd.Type,
		Cuisine:      nd.ExtraTags.Cuisine,
		OpeningHours: nd.ExtraTags.OpeningHours,
		Wheelchair:   nd.ExtraTags.Wheelchair,
		Website:      nd.ExtraTags.Website,
		Email:        nd.ExtraTags.Email,
		Phone:        nd.ExtraTags.Phone,
		Facebook:     nd.ExtraTags.Facebook,
		Instagram:    nd.ExtraTags.Instagram,
		LastOSMSync:  time.Now(),
	}
}
