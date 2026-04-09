const rtcConfiguration = {
  iceServers: [
    {
      urls: "stun:stun.l.google.com:19302",
    },
  ],
};

export class WebRTCManager {
  constructor({ localStream, sendSignal, onRemoteStream, onPeerDisconnected }) {
    this.localStream = localStream;
    this.sendSignal = sendSignal;
    this.onRemoteStream = onRemoteStream;
    this.onPeerDisconnected = onPeerDisconnected;
    this.peers = new Map();
  }

  ensurePeerConnection(peerId) {
    const existing = this.peers.get(peerId);
    if (existing) {
      return existing;
    }

    const connection = new RTCPeerConnection(rtcConfiguration);

    this.localStream.getTracks().forEach((track) => {
      connection.addTrack(track, this.localStream);
    });

    connection.onicecandidate = (event) => {
      if (!event.candidate) {
        return;
      }

      this.sendSignal({
        type: "ice_candidate",
        to: peerId,
        payload: event.candidate.toJSON(),
      });
    };

    connection.ontrack = (event) => {
      const [stream] = event.streams;
      if (stream) {
        this.onRemoteStream(peerId, stream);
      }
    };

    connection.onconnectionstatechange = () => {
      if (["closed", "disconnected", "failed"].includes(connection.connectionState)) {
        this.removePeer(peerId);
        this.onPeerDisconnected(peerId);
      }
    };

    this.peers.set(peerId, connection);
    return connection;
  }

  async createOffer(peerId) {
    const connection = this.ensurePeerConnection(peerId);
    const offer = await connection.createOffer();
    await connection.setLocalDescription(offer);

    this.sendSignal({
      type: "offer",
      to: peerId,
      payload: connection.localDescription,
    });
  }

  async handleOffer(peerId, payload) {
    const connection = this.ensurePeerConnection(peerId);
    await connection.setRemoteDescription(new RTCSessionDescription(payload));
    const answer = await connection.createAnswer();
    await connection.setLocalDescription(answer);

    this.sendSignal({
      type: "answer",
      to: peerId,
      payload: connection.localDescription,
    });
  }

  async handleAnswer(peerId, payload) {
    const connection = this.ensurePeerConnection(peerId);
    await connection.setRemoteDescription(new RTCSessionDescription(payload));
  }

  async handleIceCandidate(peerId, payload) {
    const connection = this.ensurePeerConnection(peerId);
    await connection.addIceCandidate(new RTCIceCandidate(payload));
  }

  removePeer(peerId) {
    const connection = this.peers.get(peerId);
    if (!connection) {
      return;
    }

    connection.close();
    this.peers.delete(peerId);
  }

  closeAll() {
    for (const peerId of this.peers.keys()) {
      this.removePeer(peerId);
    }
  }
}
