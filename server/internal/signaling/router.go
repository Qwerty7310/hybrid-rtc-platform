package signaling

import (
	"encoding/json"
	"errors"
	"time"

	"hybrid-rtc-platform/server/internal/models"
	"hybrid-rtc-platform/server/internal/rooms"
)

type Router struct {
	rooms *rooms.Manager
}

func NewRouter(roomManager *rooms.Manager) *Router {
	return &Router{rooms: roomManager}
}

func (r *Router) Handle(client *models.Client, message models.Message) error {
	switch message.Type {
	case models.TypeJoinRoom:
		return r.handleJoin(client, message)
	case models.TypeOffer, models.TypeAnswer, models.TypeICECandidate:
		return r.forwardPeerMessage(client, message)
	case models.TypeChatMessage:
		return r.handleChatMessage(client, message)
	default:
		return r.sendError(client, "unsupported message type")
	}
}

func (r *Router) Disconnect(client *models.Client) {
	if client.RoomID == "" {
		_ = client.Close()
		return
	}

	room, ok := r.rooms.Get(client.RoomID)
	if !ok {
		_ = client.Close()
		return
	}

	room.RemoveClient(client.ID)
	r.sendSystemMessage(room, client.Name+" left the room")
	r.broadcastPresence(room, models.TypeUserLeft, client, client.ID)
	r.rooms.DeleteIfEmpty(client.RoomID)
	_ = client.Close()
}

func (r *Router) handleJoin(client *models.Client, message models.Message) error {
	var payload models.JoinPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return r.sendError(client, "invalid join payload")
	}

	if payload.RoomID == "" || payload.UserID == "" {
		return r.sendError(client, "roomId and userId are required")
	}

	if client.RoomID != "" {
		return r.sendError(client, "client already joined a room")
	}

	client.ID = payload.UserID
	client.Name = payload.Name
	client.RoomID = payload.RoomID
	if client.Name == "" {
		client.Name = client.ID
	}

	room, err := r.rooms.Join(payload.RoomID, client)
	if err != nil {
		client.ID = ""
		client.Name = ""
		client.RoomID = ""
		return r.sendError(client, err.Error())
	}

	joinedPayload, _ := json.Marshal(models.RoomJoinedPayload{
		RoomID: room.ID,
		CurrentUser: models.Participant{
			UserID: client.ID,
			Name:   client.Name,
		},
		Participants:    room.Participants(client.ID),
		MaxParticipants: rooms.MaxParticipants,
	})
	client.SafeWriteJSON(models.MustMarshalMessage(models.Message{
		Type:    models.TypeRoomJoined,
		RoomID:  room.ID,
		From:    "server",
		Payload: joinedPayload,
	}))

	r.broadcastPresence(room, models.TypeUserJoined, client, client.ID)
	r.sendSystemMessage(room, client.Name+" joined the room")
	return nil
}

func (r *Router) handleChatMessage(client *models.Client, message models.Message) error {
	if client.RoomID == "" {
		return r.sendError(client, "join a room before sending messages")
	}

	room, ok := r.rooms.Get(client.RoomID)
	if !ok {
		return errors.New("room not found")
	}

	var payload models.ChatPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return r.sendError(client, "invalid chat payload")
	}
	if payload.Message == "" {
		return r.sendError(client, "message is required")
	}
	if payload.Timestamp == "" {
		payload.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	body, _ := json.Marshal(payload)
	broadcast := models.MustMarshalMessage(models.Message{
		Type:    models.TypeChatMessage,
		From:    client.ID,
		RoomID:  client.RoomID,
		Payload: body,
	})

	for _, peer := range room.SnapshotClients("") {
		peer.SafeWriteJSON(broadcast)
	}

	return nil
}

func (r *Router) forwardPeerMessage(client *models.Client, message models.Message) error {
	if client.RoomID == "" {
		return r.sendError(client, "join a room before signaling")
	}
	if message.To == "" {
		return r.sendError(client, "target peer is required")
	}

	room, ok := r.rooms.Get(client.RoomID)
	if !ok {
		return r.sendError(client, "room not found")
	}

	target, ok := room.GetClient(message.To)
	if !ok {
		return r.sendError(client, "target peer is not connected")
	}

	message.From = client.ID
	message.RoomID = client.RoomID
	if !target.SafeWriteJSON(models.MustMarshalMessage(message)) {
		return r.sendError(client, "target peer queue is full")
	}

	return nil
}

func (r *Router) broadcastPresence(room *models.Room, messageType string, user *models.Client, excludeID string) {
	payload, _ := json.Marshal(models.PresencePayload{
		User: models.Participant{
			UserID: user.ID,
			Name:   user.Name,
		},
	})
	message := models.MustMarshalMessage(models.Message{
		Type:    messageType,
		RoomID:  room.ID,
		From:    user.ID,
		Payload: payload,
	})

	for _, peer := range room.SnapshotClients(excludeID) {
		peer.SafeWriteJSON(message)
	}
}

func (r *Router) sendSystemMessage(room *models.Room, text string) {
	payload, _ := json.Marshal(models.SystemPayload{Message: text})
	message := models.MustMarshalMessage(models.Message{
		Type:    models.TypeSystem,
		RoomID:  room.ID,
		From:    "server",
		Payload: payload,
	})

	for _, peer := range room.SnapshotClients("") {
		peer.SafeWriteJSON(message)
	}
}

func (r *Router) sendError(client *models.Client, message string) error {
	payload, _ := json.Marshal(models.ErrorPayload{Message: message})
	client.SafeWriteJSON(models.MustMarshalMessage(models.Message{
		Type:    models.TypeError,
		Payload: payload,
	}))
	return errors.New(message)
}
