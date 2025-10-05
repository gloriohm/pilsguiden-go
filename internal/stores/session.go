package stores

import (
	"fmt"
	"go-router/models"
	"time"

	"github.com/patrickmn/go-cache"
)

func GetSessionStore(store *cache.Cache, sessionID string) models.SessionStore {
	if sess, found := store.Get(sessionID); found {
		if data, ok := sess.(models.SessionStore); ok {
			return data
		}
	}

	return models.SessionStore{}
}

func SetSessionStore(store *cache.Cache, sessionID string, data models.SessionStore) {
	store.Set(sessionID, data, cache.DefaultExpiration)
	fmt.Printf("session stored with id %s and data %v \n", sessionID, data)
}

func SetNavData(store *cache.Cache, sessionID string, data models.Navigation) {
	sess := GetSessionStore(store, sessionID)
	sess.Navigation = data
	SetSessionStore(store, sessionID, sess)
}

func SetSessionPrefs(store *cache.Cache, sessionID string, data models.Preferences) {
	sess := GetSessionStore(store, sessionID)
	sess.Preferences = data
	SetSessionStore(store, sessionID, sess)
}

func SetSessionFilter(store *cache.Cache, sessionID string, data models.BarsFilter) {
	sess := GetSessionStore(store, sessionID)
	sess.BarsFilter = data
	SetSessionStore(store, sessionID, sess)
}

func GetClosestDate(day int) string {
	// Get today's date
	today := time.Now()

	// Get the current day of the week (0–6, Sunday=0)
	currentDay := int(today.Weekday())

	// Calculate difference in days
	diff := (day - currentDay + 7) % 7

	// Add the difference to today's date
	closestDate := today.AddDate(0, 0, diff)

	// Format as YYYY-MM-DD
	return closestDate.Format("2006-01-02")
}
