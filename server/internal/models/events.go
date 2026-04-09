package models

type JoinPayload struct {
	RoomID string `json:"roomId"`
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

type Participant struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

type RoomJoinedPayload struct {
	RoomID          string        `json:"roomId"`
	CurrentUser     Participant   `json:"currentUser"`
	Participants    []Participant `json:"participants"`
	MaxParticipants int           `json:"maxParticipants"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}

type PresencePayload struct {
	User Participant `json:"user"`
}

type ChatPayload struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type SystemPayload struct {
	Message string `json:"message"`
}
