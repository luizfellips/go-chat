import { create } from "zustand";

type WSStatus = "disconnected" | "connecting" | "connected";

interface WebSocketState {
  status: WSStatus;
  setStatus: (status: WSStatus) => void;
}

export const useWebSocketStore = create<WebSocketState>((set) => ({
  status: "disconnected",
  setStatus: (status) => set({ status }),
}));
