import { baseURL } from "@/utils/constants";
import { useAuthStore } from "@/stores/auth";

type ChangeCallback = (data: { path: string; op: string }) => void;

class LiveSyncClient {
  private ws: WebSocket | null = null;
  private currentPath: string = "";
  private listeners: Set<ChangeCallback> = new Set();
  private reconnectTimer: number | null = null;

  public connect() {
    const authStore = useAuthStore();
    if (!authStore.isLoggedIn) return;

    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return;
    }

    const loc = window.location;
    const protocol = loc.protocol === "https:" ? "wss:" : "ws:";
    const host = loc.host;
    const wsURL = `${protocol}//${host}${baseURL}/api/ws`;

    try {
      this.ws = new WebSocket(wsURL);

      this.ws.onopen = () => {
        if (this.currentPath) {
          this.sendWatch(this.currentPath);
        }
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === "change") {
            this.listeners.forEach((cb) => cb(data));
          }
        } catch {
          // ignore non-JSON messages
        }
      };

      this.ws.onclose = () => {
        this.scheduleReconnect();
      };

      this.ws.onerror = () => {
        this.ws?.close();
      };
    } catch {
      this.scheduleReconnect();
    }
  }

  private scheduleReconnect() {
    if (this.reconnectTimer) return;
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, 5000);
  }

  private sendWatch(path: string) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ action: "watch", path }));
    }
  }

  public watch(path: string, callback: ChangeCallback): () => void {
    this.currentPath = path;
    this.listeners.add(callback);
    this.connect();
    this.sendWatch(path);

    return () => {
      this.listeners.delete(callback);
    };
  }

  public disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

export const liveSync = new LiveSyncClient();
