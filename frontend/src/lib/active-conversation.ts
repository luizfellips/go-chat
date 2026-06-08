let activeConversationId: string | null = null;

export function setActiveConversationId(id: string | null) {
  activeConversationId = id;
}

export function getActiveConversationId() {
  return activeConversationId;
}
