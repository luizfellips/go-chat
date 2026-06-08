import { create } from "zustand";
import type { User } from "@/types/auth";
import { clearTokens, getRefreshToken, setTokens } from "@/lib/token";

interface AuthState {
  user: User | null;
  accessToken: string | null;
  setAuth: (user: User, accessToken: string, refreshToken: string) => void;
  setUser: (user: User) => void;
  setAccessToken: (token: string) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  accessToken: null,
  setAuth: (user, accessToken, refreshToken) => {
    setTokens(accessToken, refreshToken);
    set({ user, accessToken });
  },
  setUser: (user) => set({ user }),
  setAccessToken: (token) => {
    const refresh = getRefreshToken();
    if (refresh) setTokens(token, refresh);
    set({ accessToken: token });
  },
  logout: () => {
    clearTokens();
    set({ user: null, accessToken: null });
  },
}));
