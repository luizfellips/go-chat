import { api } from "./api";
import type { Conversation } from "@/types/conversation";

export async function listConversations(): Promise<Conversation[]> {
  const { data } = await api.get<{ conversations: Conversation[] }>("/conversations");
  return data.conversations;
}

export async function createConversation(participantId: string): Promise<{ id: string }> {
  const { data } = await api.post("/conversations", { participant_id: participantId });
  return data;
}
