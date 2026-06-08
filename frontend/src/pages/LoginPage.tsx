import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { login } from "@/services/auth.service";
import { useAuthStore } from "@/store/auth.store";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ThemeToggle } from "@/components/layout/ThemeToggle";

export function LoginPage() {
  const [email, setEmail] = useState("alice@example.com");
  const [password, setPassword] = useState("password123");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const setAuth = useAuthStore((s) => s.setAuth);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const data = await login(email, password);
      setAuth(data.user, data.access_token, data.refresh_token);
      navigate("/");
    } catch {
      setError("Invalid credentials");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <div className="absolute right-4 top-4">
        <ThemeToggle />
      </div>
      <form onSubmit={handleSubmit} className="w-full max-w-md space-y-4 rounded-xl border border-border bg-card p-8 shadow-sm">
        <h1 className="text-2xl font-bold">Go Chat</h1>
        <p className="text-sm text-muted-foreground">Sign in to continue</p>
        {error && <p className="text-sm text-red-500">{error}</p>}
        <Input type="email" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        <Input type="password" placeholder="Password" value={password} onChange={(e) => setPassword(e.target.value)} required />
        <Button type="submit" className="w-full" disabled={loading}>
          {loading ? "Signing in..." : "Sign in"}
        </Button>
        <p className="text-center text-sm text-muted-foreground">
          No account? <Link to="/register" className="text-primary hover:underline">Register</Link>
        </p>
        <p className="text-xs text-muted-foreground">
          Demo: alice@example.com / password123 or bob@example.com / password123
        </p>
      </form>
    </div>
  );
}
