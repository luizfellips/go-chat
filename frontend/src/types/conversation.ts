export interface ConversationParticipant {
  id: string;
  username: string;
  is_online: boolean;
}

export interface LastMessage {
  content?: string | null;
  created_at?: string | null;
  sender_id?: string | null;
}

export interface Conversation {
  id: string;
  participant: ConversationParticipant;
  last_message?: LastMessage | null;
  unread_count: number;
}
