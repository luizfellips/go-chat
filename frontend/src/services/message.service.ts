import { api } from "./api";
import type { Message, MessagesResponse } from "@/types/message";

export async function listMessages(
  conversationId: string,
  cursor?: string,
  limit = 50
): Promise<MessagesResponse> {
  const params = new URLSearchParams({ limit: String(limit) });
  if (cursor) params.set("cursor", cursor);
  const { data } = await api.get<MessagesResponse>(
    `/conversations/${conversationId}/messages?${params}`
  );
  return data;
}

export async function sendMessage(conversationId: string, content: string): Promise<Message> {
  const { data } = await api.post<Message>(`/conversations/${conversationId}/messages`, {
    content,
  });
  return data;
}
