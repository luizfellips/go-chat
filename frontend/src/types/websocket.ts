export type WSEventType =
  | "connection"
  | "disconnect"
  | "message_sent"
  | "message_received"
  | "typing_start"
  | "typing_stop"
  | "user_online"
  | "user_offline"
  | "message_read";

export interface WSEnvelope<T = unknown> {
  type: WSEventType;
  payload: T;
  timestamp: string;
}

export interface WSMessageReceivedPayload {
  message: {
    id: string;
    conversation_id: string;
    sender_id?: string | null;
    content: string;
    created_at: string;
    read_at?: string | null;
  };
}

export interface WSMessageReadPayload {
  conversation_id: string;
  message_id: string;
  read_at: string;
}

export interface WSUserPresencePayload {
  user_id: string;
}

export interface WSTypingPayload {
  conversation_id: string;
  user_id: string;
}
