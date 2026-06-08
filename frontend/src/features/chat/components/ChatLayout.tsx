import { ReactNode } from "react";
import { ThemeToggle } from "@/components/layout/ThemeToggle";
import { useAuthStore } from "@/store/auth.store";
import { logout } from "@/services/auth.service";
import { Button } from "@/components/ui/button";
import { useNavigate } from "react-router-dom";

interface ChatLayoutProps {
  sidebar: ReactNode;
  main: ReactNode;
  showMobileChat?: boolean;
}

export function ChatLayout({ sidebar, main, showMobileChat = false }: ChatLayoutProps) {
  const user = useAuthStore((s) => s.user);
  const logoutStore = useAuthStore((s) => s.logout);
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    logoutStore();
    navigate("/login");
  };

  return (
    <div className="flex h-dvh min-h-0 flex-col bg-background">
      <header className="flex shrink-0 items-center justify-between border-b border-border px-3 py-2 sm:px-4 sm:py-3">
        <div className="min-w-0">
          <h1 className="truncate text-base font-bold sm:text-lg">Go Chat</h1>
          {user && (
            <p className="truncate text-xs text-muted-foreground">
              <span className="hidden sm:inline">Signed in as </span>
              {user.username}
            </p>
          )}
        </div>
        <div className="flex shrink-0 items-center gap-1 sm:gap-2">
          <ThemeToggle />
          <Button variant="outline" size="sm" onClick={handleLogout} className="px-2 sm:px-3">
            <span className="hidden sm:inline">Logout</span>
            <span className="sm:hidden">Exit</span>
          </Button>
        </div>
      </header>

      <div className="flex min-h-0 flex-1 overflow-hidden">
        <div
          className={`h-full min-h-0 w-full shrink-0 md:w-80 lg:w-96 ${
            showMobileChat ? "hidden md:flex" : "flex"
          }`}
        >
          {sidebar}
        </div>
        <main
          className={`min-h-0 min-w-0 flex-1 flex-col bg-chat ${
            showMobileChat ? "flex" : "hidden md:flex"
          }`}
        >
          {main}
        </main>
      </div>
    </div>
  );
}
