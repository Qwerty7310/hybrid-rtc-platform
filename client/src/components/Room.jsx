import Chat from "./Chat";
import VideoGrid from "./VideoGrid";

export default function Room({
  roomState,
  localStream,
  remoteStreams,
  feedItems,
  onSendChat,
  onLeave,
}) {
  return (
    <main className="room-layout">
      <section className="panel room-panel">
        <header className="room-header">
          <div>
            <h1>Room: {roomState.roomId}</h1>
            <p className="muted">
              Participants: {roomState.participants.length}/{roomState.maxParticipants}
            </p>
          </div>
          <button className="button-secondary" onClick={onLeave}>
            Leave
          </button>
        </header>
        <VideoGrid
          participants={roomState.participants}
          localStream={localStream}
          remoteStreams={remoteStreams}
        />
      </section>

      <aside className="panel sidebar">
        <section className="participant-list">
          <h2>Members</h2>
          <ul>
            {roomState.participants.map((participant) => (
              <li key={participant.userId}>
                {participant.name}
                {participant.userId === roomState.currentUser?.userId ? " (you)" : ""}
              </li>
            ))}
          </ul>
        </section>
        <Chat
          feedItems={feedItems}
          onSend={onSendChat}
        />
      </aside>
    </main>
  );
}
