package models

import "time"

type Preferences struct {
	CustomTime bool
	Time       time.Time
	Date       string
}

type Navigation struct {
	Level string
	ID    int
}

type SessionData struct {
	Navigation  Navigation
	Preferences Preferences
}

type RawCustomTime struct {
	Time string `form:"customTime"`
	Day  int    `form:"customDay"`
}
