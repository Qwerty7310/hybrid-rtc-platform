export class WebSocketService {
  constructor(url) {
    this.url = url;
    this.socket = null;
    this.messageHandler = null;
    this.closeHandler = null;
  }

  connect() {
    return new Promise((resolve, reject) => {
      this.socket = new WebSocket(this.url);

      this.socket.onopen = () => resolve();
      this.socket.onerror = () => reject(new Error("Failed to connect to WebSocket"));
      this.socket.onmessage = (event) => {
        if (!this.messageHandler) {
          return;
        }

        const message = JSON.parse(event.data);
        this.messageHandler(message);
      };
      this.socket.onclose = () => {
        if (this.closeHandler) {
          this.closeHandler();
        }
      };
    });
  }

  onMessage(handler) {
    this.messageHandler = handler;
  }

  onClose(handler) {
    this.closeHandler = handler;
  }

  send(type, payload, to = "") {
    this.socket?.send(
      JSON.stringify({
        type,
        to,
        payload,
      }),
    );
  }

  close() {
    if (!this.socket || this.socket.readyState === WebSocket.CLOSED) {
      return;
    }

    this.socket.close();
  }
}
