import { create } from "zustand";

interface PresenceState {
  online: Record<string, boolean>;
  setOnline: (userId: string, isOnline: boolean) => void;
  isOnline: (userId: string) => boolean;
}

export const usePresenceStore = create<PresenceState>((set, get) => ({
  online: {},
  setOnline: (userId, isOnline) =>
    set((s) => ({ online: { ...s.online, [userId]: isOnline } })),
  isOnline: (userId) => get().online[userId] ?? false,
}));
