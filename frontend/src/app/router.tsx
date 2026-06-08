import { createBrowserRouter, Navigate } from "react-router-dom";
import { ProtectedRoute } from "@/lib/ProtectedRoute";
import { LoginPage } from "@/pages/LoginPage";
import { RegisterPage } from "@/pages/RegisterPage";
import { ChatPage } from "@/pages/ChatPage";
import { getMe } from "@/services/auth.service";
import { getAccessToken } from "@/lib/token";
import { useAuthStore } from "@/store/auth.store";

async function chatLoader() {
  const token = getAccessToken();
  if (!token) return null;
  const store = useAuthStore.getState();
  if (!store.user) {
    try {
      const user = await getMe();
      store.setUser(user);
    } catch {
      return null;
    }
  }
  return null;
}

export const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  { path: "/register", element: <RegisterPage /> },
  {
    element: <ProtectedRoute />,
    loader: chatLoader,
    children: [{ path: "/", element: <ChatPage /> }],
  },
  { path: "*", element: <Navigate to="/" replace /> },
]);
