package rooms

import (
	"errors"
	"sync"

	"hybrid-rtc-platform/server/internal/models"
)

const MaxParticipants = 4

type Manager struct {
	mu    sync.RWMutex
	Rooms map[string]*models.Room
}

func NewManager() *Manager {
	return &Manager{
		Rooms: make(map[string]*models.Room),
	}
}

func (m *Manager) Get(roomID string) (*models.Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	room, ok := m.Rooms[roomID]
	return room, ok
}

func (m *Manager) GetOrCreate(roomID string) *models.Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.Rooms[roomID]
	if ok {
		return room
	}

	room = models.NewRoom(roomID)
	m.Rooms[roomID] = room
	return room
}

func (m *Manager) DeleteIfEmpty(roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.Rooms[roomID]
	if !ok || !room.IsEmpty() {
		return
	}

	delete(m.Rooms, roomID)
}

func (m *Manager) Join(roomID string, client *models.Client) (*models.Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.Rooms[roomID]
	if !ok {
		room = models.NewRoom(roomID)
		m.Rooms[roomID] = room
	}

	if room.ParticipantCount() >= MaxParticipants {
		return nil, errors.New("room is full")
	}

	if _, exists := room.GetClient(client.ID); exists {
		return nil, errors.New("userId already exists in room")
	}

	room.AddClient(client)
	return room, nil
}
