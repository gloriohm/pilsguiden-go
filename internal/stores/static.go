package stores

import (
	"errors"
	"go-router/models"
	"sync"
	"time"
)

type Fylker struct {
	Data        []models.Location
	LastUpdated time.Time
}

type Kommuner struct {
	Data        []models.Location
	LastUpdated time.Time
}

type Steder struct {
	Data        []models.Location
	LastUpdated time.Time
}

type Breweries struct {
	Data        []models.Brewery
	LastUpdated time.Time
}

type StaticStore struct {
	mu        sync.RWMutex
	Fylker    Fylker
	Kommuner  Kommuner
	Steder    Steder
	Breweries Breweries
}

var AppStore = &StaticStore{}

func (s *StaticStore) UpdateFylker(data []models.Location) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Fylker = Fylker{
		Data:        data,
		LastUpdated: time.Now(),
	}
}

func (s *StaticStore) UpdateKommuner(data []models.Location) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Kommuner = Kommuner{
		Data:        data,
		LastUpdated: time.Now(),
	}
}

func (s *StaticStore) UpdateSteder(data []models.Location) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Steder = Steder{
		Data:        data,
		LastUpdated: time.Now(),
	}
}

func (s *StaticStore) UpdateBreweries(data []models.Brewery) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Breweries = Breweries{
		Data:        data,
		LastUpdated: time.Now(),
	}
}

func (s *StaticStore) GetFylkerData() []models.Location {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Fylker.Data
}

func (s *StaticStore) GetFylkeBySlug(slug string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, fylke := range s.Fylker.Data {
		if fylke.Slug == slug {
			return fylke.ID
		}
	}
	return 0
}

func (s *StaticStore) GetKommunerData() []models.Location {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Kommuner.Data
}

func (s *StaticStore) GetKommuneBySlug(slug string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, kommune := range s.Kommuner.Data {
		if kommune.Slug == slug {
			return kommune.ID
		}
	}
	return 0
}

func (s *StaticStore) GetStederData() []models.Location {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Steder.Data
}

func (s *StaticStore) GetStedBySlug(slug string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sted := range s.Steder.Data {
		if sted.Slug == slug {
			return sted.ID
		}
	}
	return 0
}

func (s *StaticStore) GetBreweriesData() []models.Brewery {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Breweries.Data
}

func (s *StaticStore) GetLocationBySlug(slug, level string) (models.BaseLocation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var src []models.Location
	switch level {
	case "fylke":
		src = s.Fylker.Data
	case "kommune":
		src = s.Kommuner.Data
	case "sted":
		src = s.Steder.Data
	default:
		return models.BaseLocation{}, errors.New("level not fylke, kommune, or sted")
	}

	for _, loc := range src {
		if loc.Slug == slug {
			return models.BaseLocation{ID: loc.ID, Name: loc.Name, Slug: loc.Slug}, nil
		}
	}

	return models.BaseLocation{}, errors.New("location not found by slug")
}

func (s *StaticStore) GetLocationsByParent(ID int, level string) []models.Location {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var src []models.Location
	switch level {
	case "kommune":
		src = s.Kommuner.Data
	case "sted":
		src = s.Steder.Data
	default:
		return nil
	}

	var matches []models.Location
	for _, loc := range src {
		if *loc.Parent == ID {
			matches = append(matches, loc)
		}
	}
	return matches
}

func (s *StaticStore) BreweryInBreweries(target string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.Breweries.Data {
		if item.Name == target {
			return true
		}
	}
	return false
}
