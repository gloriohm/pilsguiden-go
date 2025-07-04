package models

import "time"

type Preferences struct {
	CustomTime bool
	Time       time.Time
	Date       string
}

type SessionData struct {
	Preferences Preferences
}
