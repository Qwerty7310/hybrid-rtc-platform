import { useState } from "react";

export default function JoinForm({ onJoin, isJoining }) {
  const [name, setName] = useState("");
  const [roomId, setRoomId] = useState("");

  function handleSubmit(event) {
    event.preventDefault();
    onJoin({
      name: name.trim(),
      roomId: roomId.trim(),
    });
  }

  return (
    <form className="join-form" onSubmit={handleSubmit}>
      <label>
        <span>Name</span>
        <input
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder="Alice"
          minLength={2}
          maxLength={24}
          required
        />
      </label>
      <label>
        <span>Room ID</span>
        <input
          value={roomId}
          onChange={(event) => setRoomId(event.target.value)}
          placeholder="team-sync"
          minLength={2}
          maxLength={40}
          required
        />
      </label>
      <button type="submit" disabled={isJoining}>
        {isJoining ? "Joining..." : "Join room"}
      </button>
    </form>
  );
}
