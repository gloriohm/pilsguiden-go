package stores

import (
	"fmt"
	"go-router/models"
	"time"

	"github.com/patrickmn/go-cache"
)

func GetSessionData(store *cache.Cache, sessionID string) models.SessionData {
	if sess, found := store.Get(sessionID); found {
		if data, ok := sess.(models.SessionData); ok {
			return data
		}
	}

	return models.SessionData{}
}

func SetSessionData(store *cache.Cache, sessionID string, data models.SessionData) {
	store.Set(sessionID, data, cache.DefaultExpiration)
	fmt.Printf("session stored with id %s and data %v", sessionID, data)
}

func SetNavData(store *cache.Cache, sessionID string, data models.Navigation) {
	sess := GetSessionData(store, sessionID)
	sess.Navigation = data
	SetSessionData(store, sessionID, sess)
}

func SetSessionPrefs(store *cache.Cache, sessionID string, data models.Preferences) {
	sess := GetSessionData(store, sessionID)
	fmt.Println(sess)
	sess.Preferences = data
	SetSessionData(store, sessionID, sess)
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
