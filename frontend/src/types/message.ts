export interface Message {
  id: string;
  sender_id?: string | null;
  content: string;
  created_at: string;
  read_at?: string | null;
}

export interface MessagesResponse {
  messages: Message[];
  next_cursor?: string;
}
