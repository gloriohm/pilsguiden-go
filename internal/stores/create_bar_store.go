package stores

import (
	"go-router/models"
	"sync"
)

type CreateBarStore struct {
	mu          sync.RWMutex
	bar         *models.Bar
	address     *models.AddressParts
	barMetadata *models.BarMetadata
}

func (s *CreateBarStore) SetBar(bar *models.Bar) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bar = bar
}

func (s *CreateBarStore) GetBar() *models.Bar {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bar
}

func (s *CreateBarStore) SetAddress(addr *models.AddressParts) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.address = addr
}

func (s *CreateBarStore) GetAddress() *models.AddressParts {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.address
}

func (s *CreateBarStore) SetMetadata(meta *models.BarMetadata) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.barMetadata = meta
}

func (s *CreateBarStore) GetMetadata() *models.BarMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.barMetadata
}
