package room

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/shadovxw/monopoly/internal/game"
	"gorm.io/gorm"
)

// Manager holds all active rooms. Safe for concurrent access.
type Manager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
	rules *game.RuleSet
	db    *gorm.DB
}

func NewManager(rules *game.RuleSet, db *gorm.DB) *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
		rules: rules,
		db:    db,
	}
}

// Create allocates a new room and returns its join code.
func (m *Manager) Create() (*Room, error) {
	code := generateCode()
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.rooms[code]; exists {
		// collision — try once more
		code = generateCode()
	}
	r := NewRoom(code, m.rules, m.db)
	m.rooms[code] = r
	return r, nil
}

// Get returns an existing room by code.
func (m *Manager) Get(code string) (*Room, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.rooms[code]
	if !ok {
		return nil, fmt.Errorf("room %q not found", code)
	}
	return r, nil
}

// Delete removes a room from the registry.
func (m *Manager) Delete(code string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, code)
}

func generateCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
