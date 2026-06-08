import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { register } from "@/services/auth.service";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ThemeToggle } from "@/components/layout/ThemeToggle";

export function RegisterPage() {
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await register(email, username, password);
      navigate("/login");
    } catch {
      setError("Registration failed. Email or username may already exist.");
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
        <h1 className="text-2xl font-bold">Create account</h1>
        {error && <p className="text-sm text-red-500">{error}</p>}
        <Input type="email" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        <Input placeholder="Username" value={username} onChange={(e) => setUsername(e.target.value)} required />
        <Input type="password" placeholder="Password (min 8 chars)" value={password} onChange={(e) => setPassword(e.target.value)} required />
        <Button type="submit" className="w-full" disabled={loading}>
          {loading ? "Creating..." : "Register"}
        </Button>
        <p className="text-center text-sm text-muted-foreground">
          Already have an account? <Link to="/login" className="text-primary hover:underline">Sign in</Link>
        </p>
      </form>
    </div>
  );
}
