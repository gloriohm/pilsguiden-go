package stores

import (
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

func (s *StaticStore) GetKommunerData() []models.Location {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Kommuner.Data
}

func (s *StaticStore) GetStederData() []models.Location {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Steder.Data
}

func (s *StaticStore) GetBreweriesData() []models.Brewery {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Breweries.Data
}
