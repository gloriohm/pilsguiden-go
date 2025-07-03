package stores

import (
	"go-router/models"

	"github.com/patrickmn/go-cache"
)

func GetSessionData(store *cache.Cache, sessionID string) *models.SessionData {
	if sess, found := store.Get(sessionID); found {
		if data, ok := sess.(*models.SessionData); ok {
			return data
		}
	}
	return nil
}
