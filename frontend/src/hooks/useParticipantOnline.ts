import { usePresenceStore } from "@/store/presence.store";

/** Prefer live WebSocket presence; fall back to REST snapshot when unknown. */
export function useParticipantOnline(userId: string, restIsOnline = false): boolean {
  return usePresenceStore((s) =>
    userId in s.online ? s.online[userId] : restIsOnline
  );
}
