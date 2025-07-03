package models

import "time"

type SessionData struct {
	Preferences struct {
		CustomTime bool
		Time       time.Time
		Date       string
	}
}
