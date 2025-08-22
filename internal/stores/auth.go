package stores

import (
	"fmt"
	"go-router/models"

	"github.com/patrickmn/go-cache"
)

func SetUpdateBarStore(store *cache.Cache, userID string, data models.UpdateBarStore) {
	store.Set(userID, data, cache.DefaultExpiration)
	fmt.Printf("update bar store set with user %s and data %v \n", userID, data)
}

func GetUpdateBarStore(store *cache.Cache, userID string) models.UpdateBarStore {
	if user, found := store.Get(userID); found {
		if data, ok := user.(models.UpdateBarStore); ok {
			return data
		}
	}

	return models.UpdateBarStore{}
}
