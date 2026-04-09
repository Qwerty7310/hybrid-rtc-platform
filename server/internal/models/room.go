package models

import "sync"

type Room struct {
	ID      string
	Clients map[string]*Client
	mu      sync.RWMutex
}

func NewRoom(id string) *Room {
	return &Room{
		ID:      id,
		Clients: make(map[string]*Client),
	}
}

func (r *Room) AddClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Clients[client.ID] = client
}

func (r *Room) RemoveClient(clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Clients, clientID)
}

func (r *Room) GetClient(clientID string) (*Client, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	client, ok := r.Clients[clientID]
	return client, ok
}

func (r *Room) ParticipantCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Clients)
}

func (r *Room) Participants(excludeID string) []Participant {
	r.mu.RLock()
	defer r.mu.RUnlock()

	participants := make([]Participant, 0, len(r.Clients))
	for id, client := range r.Clients {
		if id == excludeID {
			continue
		}
		participants = append(participants, Participant{
			UserID: client.ID,
			Name:   client.Name,
		})
	}

	return participants
}

func (r *Room) SnapshotClients(excludeID string) []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]*Client, 0, len(r.Clients))
	for id, client := range r.Clients {
		if id == excludeID {
			continue
		}
		clients = append(clients, client)
	}

	return clients
}

func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Clients) == 0
}
