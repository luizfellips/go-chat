import { api } from "./api";
import type { LoginResponse, User } from "@/types/auth";

export async function login(email: string, password: string): Promise<LoginResponse> {
  const { data } = await api.post<LoginResponse>("/auth/login", { email, password });
  return data;
}

export async function register(
  email: string,
  username: string,
  password: string
): Promise<{ user: User }> {
  const { data } = await api.post("/auth/register", { email, username, password });
  return data;
}

export async function logout(): Promise<void> {
  const refresh = sessionStorage.getItem("gochat_refresh");
  if (refresh) {
    await api.post("/auth/logout", { refresh_token: refresh }).catch(() => {});
  }
}

export async function getMe(): Promise<User> {
  const { data } = await api.get<User>("/users/me");
  return data;
}
