package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"hybrid-rtc-platform/server/internal/models"
	"hybrid-rtc-platform/server/internal/signaling"
)

type Handler struct {
	router   *signaling.Router
	upgrader websocket.Upgrader
}

func NewHandler(router *signaling.Router) *Handler {
	return &Handler{
		router: router,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade websocket", http.StatusBadRequest)
		return
	}

	client := &models.Client{
		Conn:      conn,
		Send:      make(chan []byte, 32),
		UserAgent: r.UserAgent(),
	}

	go h.writePump(client)
	h.readPump(client)
}

func (h *Handler) readPump(client *models.Client) {
	defer h.router.Disconnect(client)

	client.Conn.SetReadLimit(models.MaxMessageSize)
	_ = client.Conn.SetReadDeadline(time.Now().Add(models.PongWait))
	client.Conn.SetPongHandler(func(string) error {
		return client.Conn.SetReadDeadline(time.Now().Add(models.PongWait))
	})

	for {
		_, data, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		var message models.Message
		if err := json.Unmarshal(data, &message); err != nil {
			log.Printf("invalid JSON message: %v", err)
			continue
		}

		if err := h.router.Handle(client, message); err != nil {
			log.Printf("client %q: %v", client.ID, err)
		}
	}
}

func (h *Handler) writePump(client *models.Client) {
	ticker := time.NewTicker(models.PingPeriod)
	defer func() {
		ticker.Stop()
		_ = client.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(models.WriteWait))
			if !ok {
				_ = client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(models.WriteWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
