package models

import "encoding/json"

type Message struct {
	Type    string          `json:"type"`
	From    string          `json:"from"`
	To      string          `json:"to"`
	RoomID  string          `json:"roomId"`
	Payload json.RawMessage `json:"payload"`
}

const (
	TypeJoinRoom     = "join_room"
	TypeRoomJoined   = "room_joined"
	TypeUserJoined   = "user_joined"
	TypeOffer        = "offer"
	TypeAnswer       = "answer"
	TypeICECandidate = "ice_candidate"
	TypeChatMessage  = "chat_message"
	TypeUserLeft     = "user_left"
	TypeSystem       = "system_message"
	TypeError        = "error"
)

func MustMarshalMessage(message Message) []byte {
	data, err := json.Marshal(message)
	if err != nil {
		fallback, _ := json.Marshal(Message{
			Type: TypeError,
			Payload: json.RawMessage(
				`{"message":"failed to encode message"}`,
			),
		})
		return fallback
	}

	return data
}
