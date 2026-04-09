# Hybrid RTC Platform

MVP веб-приложения для комнат на 3-4 участников с гибридной архитектурой:

- WebSocket-сервер на Go отвечает за подключение, комнаты, список участников, сигналинг WebRTC, текстовый чат и системные события.
- WebRTC mesh используется для прямой передачи аудио и видео между браузерами.
- Данные хранятся in-memory, без БД и TURN.

## Архитектура

### Сервер

- [server/cmd/server/main.go](/home/qwerty/hybrid-rtc-platform/server/cmd/server/main.go) поднимает HTTP/WebSocket сервер.
- [server/internal/ws/client.go](/home/qwerty/hybrid-rtc-platform/server/internal/ws/client.go) обрабатывает upgrade до WebSocket и lifecycle клиента.
- [server/internal/signaling/router.go](/home/qwerty/hybrid-rtc-platform/server/internal/signaling/router.go) маршрутизирует `join_room`, `offer`, `answer`, `ice_candidate`, `chat_message`.
- [server/internal/rooms/manager.go](/home/qwerty/hybrid-rtc-platform/server/internal/rooms/manager.go) управляет комнатами и ограничением в 4 участника.
- [server/internal/models](/home/qwerty/hybrid-rtc-platform/server/internal/models) содержит модели сообщений, payload и комнаты.

### Клиент

- [client/src/App.jsx](/home/qwerty/hybrid-rtc-platform/client/src/App.jsx) управляет сценарием входа в комнату и состоянием приложения.
- [client/src/services/websocket.js](/home/qwerty/hybrid-rtc-platform/client/src/services/websocket.js) инкапсулирует WebSocket-подключение.
- [client/src/webrtc/WebRTCManager.js](/home/qwerty/hybrid-rtc-platform/client/src/webrtc/WebRTCManager.js) управляет `RTCPeerConnection` для каждого участника.
- [client/src/components/Room.jsx](/home/qwerty/hybrid-rtc-platform/client/src/components/Room.jsx), [VideoGrid.jsx](/home/qwerty/hybrid-rtc-platform/client/src/components/VideoGrid.jsx) и [Chat.jsx](/home/qwerty/hybrid-rtc-platform/client/src/components/Chat.jsx) дают упрощённый UI.

## Протокол сигналинга

Сообщения идут через WebSocket в формате:

```json
{
  "type": "offer",
  "from": "user-a",
  "to": "user-b",
  "roomId": "team-sync",
  "payload": {}
}
```

Используемые типы:

- `join_room`
- `room_joined`
- `user_joined`
- `offer`
- `answer`
- `ice_candidate`
- `chat_message`
- `user_left`
- `system_message`
- `error`

## Запуск

### 1. Сервер

```bash
cd server
go run ./cmd/server
```

По умолчанию сервер слушает `:8080`, WebSocket endpoint: `ws://localhost:8080/ws`.

### 2. Клиент

```bash
cd client
npm install
npm run dev
```

При необходимости можно переопределить адрес сигналинга:

```bash
VITE_WS_URL=ws://localhost:8080/ws npm run dev
```

### 3. Проверка

1. Откройте приложение минимум в двух вкладках или разных браузерах.
2. Введите разные имена и один `Room ID`.
3. Разрешите доступ к камере и микрофону.
4. После входа второго и следующих пользователей существующие участники создадут `offer`, новый участник отправит `answer`, затем пойдёт обмен `ICE candidate`.
5. Текстовый чат идёт через WebSocket, аудио и видео через WebRTC.

## Ограничения MVP

- Комната ограничена 4 участниками.
- Используется только STUN `stun.l.google.com:19302`.
- TURN, авторизация, запись, persistent storage и production-hardening не входят в этот MVP.
