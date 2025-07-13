package stores

import (
	"go-router/models"
	"time"

	"github.com/patrickmn/go-cache"
)

var c = cache.New(5*time.Minute, 10*time.Minute)

func SetBarStore(sessID string, data models.BarManual) {
	c.Set(sessID, data, cache.DefaultExpiration)
}

func GetBarStore(sessionID string) models.BarManual {
	if sess, found := c.Get(sessionID); found {
		if data, ok := sess.(models.BarManual); ok {
			return data
		}
	}

	return models.BarManual{}
}

func SetAddressStore(sessID string, data models.AddressParts) {
	c.Set(sessID, data, cache.DefaultExpiration)
}

func GetAddressStore(sessionID string) models.AddressParts {
	if sess, found := c.Get(sessionID); found {
		if data, ok := sess.(models.AddressParts); ok {
			return data
		}
	}

	return models.AddressParts{}
}

func SetMetaStore(sessID string, data models.BarMetadata) {
	c.Set(sessID, data, cache.DefaultExpiration)
}

func GetMetaStore(sessionID string) models.BarMetadata {
	if sess, found := c.Get(sessionID); found {
		if data, ok := sess.(models.BarMetadata); ok {
			return data
		}
	}

	return models.BarMetadata{}
}
