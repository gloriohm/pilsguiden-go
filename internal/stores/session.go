package stores

import (
	"go-router/models"
	"time"

	"github.com/patrickmn/go-cache"
)

func GetSessionData(store *cache.Cache, sessionID string) *models.SessionData {
	if sess, found := store.Get(sessionID); found {
		if data, ok := sess.(*models.SessionData); ok {
			return data
		}
	}
	return &models.SessionData{Preferences: models.Preferences{
		CustomTime: true,
		Time:       time.Now(),
		Date:       "2025-07-03",
	},
	}
}

func SetSessionData(store *cache.Cache, sessionID string, data *models.SessionData) {
	store.Set(sessionID, &data, cache.DefaultExpiration)
}
