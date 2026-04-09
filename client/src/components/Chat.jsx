import { useState } from "react";

export default function Chat({ feedItems, onSend }) {
  const [message, setMessage] = useState("");

  function handleSubmit(event) {
    event.preventDefault();
    onSend(message);
    setMessage("");
  }

  return (
    <section className="chat">
      <h2>Chat</h2>
      <div className="chat-feed">
        {feedItems.map((item) => (
          <div key={item.id} className={`chat-item chat-item--${item.kind}`}>
            <span>
              {item.kind === "chat" ? `${item.name}: ${item.message}` : item.message}
            </span>
            {item.timestamp ? (
              <time dateTime={item.timestamp}>
                {new Date(item.timestamp).toLocaleTimeString()}
              </time>
            ) : null}
          </div>
        ))}
      </div>
      <form className="chat-form" onSubmit={handleSubmit}>
        <input
          value={message}
          onChange={(event) => setMessage(event.target.value)}
          placeholder="Type a message"
          maxLength={500}
        />
        <button type="submit">Send</button>
      </form>
    </section>
  );
}
