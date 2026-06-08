import axios, { type AxiosError, type InternalAxiosRequestConfig } from "axios";
import { getAccessToken, getRefreshToken, setTokens, clearTokens } from "@/lib/token";
import { useAuthStore } from "@/store/auth.store";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1";

export const api = axios.create({
  baseURL: API_URL,
  headers: { "Content-Type": "application/json" },
});

api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = getAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

let refreshing: Promise<string | null> | null = null;

api.interceptors.response.use(
  (res) => res,
  async (error: AxiosError) => {
    const original = error.config;
    if (!original || error.response?.status !== 401 || original.url?.includes("/auth/")) {
      return Promise.reject(error);
    }
    if (!refreshing) {
      refreshing = (async () => {
        const refresh = getRefreshToken();
        if (!refresh) return null;
        try {
          const { data } = await axios.post(`${API_URL}/auth/refresh`, {
            refresh_token: refresh,
          });
          setTokens(data.access_token, data.refresh_token);
          useAuthStore.getState().setAccessToken(data.access_token);
          return data.access_token as string;
        } catch {
          clearTokens();
          return null;
        } finally {
          refreshing = null;
        }
      })();
    }
    const newToken = await refreshing;
    if (!newToken) return Promise.reject(error);
    original.headers.Authorization = `Bearer ${newToken}`;
    return api(original);
  }
);
