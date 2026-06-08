import { api } from "./api";
import type { User } from "@/types/auth";

export async function searchUserByUsername(username: string): Promise<User> {
  const { data } = await api.get<User>(`/users/search?username=${encodeURIComponent(username)}`);
  return data;
}
