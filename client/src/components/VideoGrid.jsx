import { useEffect, useRef } from "react";

function VideoTile({ label, stream, muted = false }) {
  const videoRef = useRef(null);

  useEffect(() => {
    if (!videoRef.current) {
      return;
    }

    videoRef.current.srcObject = stream ?? null;
  }, [stream]);

  return (
    <article className="video-tile">
      <video ref={videoRef} autoPlay playsInline muted={muted} />
      <div className="video-caption">{label}</div>
    </article>
  );
}

export default function VideoGrid({ participants, localStream, remoteStreams }) {
  return (
    <section className="video-grid">
      <VideoTile label="You" stream={localStream} muted />
      {participants
        .filter((participant) => remoteStreams[participant.userId])
        .map((participant) => (
          <VideoTile
            key={participant.userId}
            label={participant.name}
            stream={remoteStreams[participant.userId]}
          />
        ))}
    </section>
  );
}
