import { useEffect, useRef, useState } from "react";
import JoinForm from "./components/JoinForm";
import Room from "./components/Room";
import { WebSocketService } from "./services/websocket";
import { WebRTCManager } from "./webrtc/WebRTCManager";

function getDefaultWsUrl() {
  const basePath = import.meta.env.BASE_URL || "/";
  const normalizedBasePath = basePath.endsWith("/") ? basePath.slice(0, -1) : basePath;
  const wsPath = `${normalizedBasePath || ""}/ws/`;
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}${wsPath}`;
}

const wsUrl = import.meta.env.VITE_WS_URL ?? getDefaultWsUrl();

function makeUserId() {
  return `user-${Math.random().toString(36).slice(2, 10)}`;
}

export default function App() {
  const [view, setView] = useState("lobby");
  const [error, setError] = useState("");
  const [roomState, setRoomState] = useState(null);
  const [remoteStreams, setRemoteStreams] = useState({});
  const [feedItems, setFeedItems] = useState([]);
  const [isJoining, setIsJoining] = useState(false);

  const wsRef = useRef(null);
  const rtcRef = useRef(null);
  const localStreamRef = useRef(null);
  const roomStateRef = useRef(null);
  const isClosingRef = useRef(false);

  useEffect(() => {
    return () => {
      teardownSession();
    };
  }, []);

  useEffect(() => {
    roomStateRef.current = roomState;
  }, [roomState]);

  function attachWebSocketHandlers(ws, userId) {
    ws.onMessage((message) => {
      switch (message.type) {
        case "room_joined": {
          const { roomId, currentUser, participants, maxParticipants } = message.payload;
          const nextRoomState = {
            roomId,
            currentUser,
            participants: [currentUser, ...participants],
            maxParticipants,
          };
          roomStateRef.current = nextRoomState;
          setRoomState(nextRoomState);
          setFeedItems([]);
          setView("room");
          break;
        }
        case "user_joined": {
          const joinedUser = message.payload.user;
          setRoomState((current) => {
            if (!current) {
              return current;
            }

            if (current.participants.some((item) => item.userId === joinedUser.userId)) {
              return current;
            }

            return {
              ...current,
              participants: [...current.participants, joinedUser],
            };
          });

          if (joinedUser.userId !== userId) {
            rtcRef.current?.createOffer(joinedUser.userId);
          }
          break;
        }
        case "user_left": {
          const leftUser = message.payload.user;
          setRoomState((current) => {
            if (!current) {
              return current;
            }

            return {
              ...current,
              participants: current.participants.filter((item) => item.userId !== leftUser.userId),
            };
          });

          setRemoteStreams((current) => {
            const next = { ...current };
            delete next[leftUser.userId];
            return next;
          });

          rtcRef.current?.removePeer(leftUser.userId);
          break;
        }
        case "offer":
          rtcRef.current?.handleOffer(message.from, message.payload);
          break;
        case "answer":
          rtcRef.current?.handleAnswer(message.from, message.payload);
          break;
        case "ice_candidate":
          rtcRef.current?.handleIceCandidate(message.from, message.payload);
          break;
        case "chat_message": {
          const author = roomStateRef.current?.participants.find((item) => item.userId === message.from);
          setFeedItems((current) => [...current, {
            id: `${message.from}-${message.payload.timestamp}`,
            kind: "chat",
            name: author?.name ?? message.from,
            message: message.payload.message,
            timestamp: message.payload.timestamp,
          }]);
          break;
        }
        case "system_message":
          setFeedItems((current) => [...current, {
            id: `system-${Date.now()}-${current.length}`,
            kind: "system",
            message: message.payload.message,
            timestamp: new Date().toISOString(),
          }]);
          break;
        case "error":
          setError(message.payload.message);
          break;
        default:
          break;
      }
    });

    ws.onClose(() => {
      if (isClosingRef.current) {
        isClosingRef.current = false;
        return;
      }

      teardownSession({ keepError: true });
      setView("lobby");
      setError("WebSocket connection closed");
    });
  }

  async function joinRoom({ roomId, name }) {
    setError("");
    setIsJoining(true);

    try {
      if (!roomId.trim() || !name.trim()) {
        throw new Error("Name and room ID are required");
      }

      const userId = makeUserId();
      const localStream = await navigator.mediaDevices.getUserMedia({
        audio: true,
        video: true,
      });

      const ws = new WebSocketService(wsUrl);
      await ws.connect();

      const rtc = new WebRTCManager({
        localStream,
        sendSignal: ({ type, to, payload }) => ws.send(type, payload, to),
        onRemoteStream: (peerId, stream) => {
          setRemoteStreams((current) => ({
            ...current,
            [peerId]: stream,
          }));
        },
        onPeerDisconnected: (peerId) => {
          setRemoteStreams((current) => {
            const next = { ...current };
            delete next[peerId];
            return next;
          });
        },
      });

      localStreamRef.current = localStream;
      wsRef.current = ws;
      rtcRef.current = rtc;

      attachWebSocketHandlers(ws, userId);
      ws.send("join_room", { roomId, userId, name });
    } catch (joinError) {
      teardownSession({ keepError: true });
      setError(joinError.message || "Failed to join room");
    } finally {
      setIsJoining(false);
    }
  }

  function sendChatMessage(text) {
    const trimmed = text.trim();
    if (!trimmed || !wsRef.current) {
      return;
    }

    wsRef.current.send("chat_message", {
      message: trimmed,
      timestamp: new Date().toISOString(),
    });
  }

  function leaveRoom() {
    teardownSession();
    setView("lobby");
    setRoomState(null);
    roomStateRef.current = null;
    setFeedItems([]);
    setRemoteStreams({});
    setError("");
  }

  function teardownSession(options = {}) {
    rtcRef.current?.closeAll();
    rtcRef.current = null;

    if (wsRef.current) {
      isClosingRef.current = true;
      wsRef.current.close();
    }
    wsRef.current = null;

    if (localStreamRef.current) {
      localStreamRef.current.getTracks().forEach((track) => track.stop());
      localStreamRef.current = null;
    }

    setRemoteStreams({});
    if (!options.keepError) {
      setError("");
    }
  }

  if (view === "room" && roomState) {
    return (
      <Room
        roomState={roomState}
        localStream={localStreamRef.current}
        remoteStreams={remoteStreams}
        feedItems={feedItems}
        onSendChat={sendChatMessage}
        onLeave={leaveRoom}
      />
    );
  }

  return (
    <main className="page">
      <section className="panel panel--narrow">
        <h1>Hybrid RTC Platform</h1>
        <p className="muted">
          Mesh WebRTC for audio/video, WebSocket for signaling, room state and chat.
        </p>
        <JoinForm onJoin={joinRoom} isJoining={isJoining} />
        {error ? <p className="error">{error}</p> : null}
      </section>
    </main>
  );
}
