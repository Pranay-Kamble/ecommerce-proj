"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Eye, EyeOff, Mail, Lock, User, Loader2, Zap, ArrowRight, CheckCircle2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { authApi } from "@/lib/api";
import toast from "react-hot-toast";

const AUTH_BASE = process.env.NEXT_PUBLIC_AUTH_API_URL || "http://localhost:8080";

type Role = "buyer" | "seller";

export default function RegisterPage() {
  const router = useRouter();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<Role>("buyer");
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);

  // Password strength
  const passwordChecks = {
    length: password.length >= 8,
    upper: /[A-Z]/.test(password),
    number: /[0-9]/.test(password),
  };
  const passwordStrength = Object.values(passwordChecks).filter(Boolean).length;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      await authApi.post("/register", { name, email, password, role });
      toast.success("Account created! Check your email for the OTP.");
      router.push(`/auth/verify?email=${encodeURIComponent(email)}`);
    } catch (err: any) {
      const msg = err?.response?.data?.error || "Registration failed";
      if (msg.includes("already exists")) {
        toast.error("This email is already registered. Try logging in.");
      } else {
        toast.error(msg);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleGoogleLogin = () => {
    window.location.href = `${AUTH_BASE}/api/v1/auth/google/login`;
  };

  const strengthColors = ["", "bg-destructive", "bg-yellow-500", "bg-green-500"];
  const strengthLabels = ["", "Weak", "Fair", "Strong"];

  return (
    <div className="min-h-screen hero-gradient flex items-center justify-center px-4 py-20">
      <div className="absolute top-1/3 right-1/4 w-72 h-72 bg-primary/5 rounded-full blur-3xl pointer-events-none" />
      <div className="absolute bottom-1/3 left-1/4 w-60 h-60 bg-violet-500/5 rounded-full blur-3xl pointer-events-none" />

      <div className="relative w-full max-w-md">
        <div className="glass-strong rounded-3xl p-8 border border-border/60 shadow-2xl shadow-black/30">
          {/* Logo */}
          <div className="flex justify-center mb-6">
            <Link href="/" className="flex items-center gap-2 group">
              <div className="w-10 h-10 rounded-xl bg-primary flex items-center justify-center glow-primary transition-transform group-hover:scale-110">
                <Zap className="w-5 h-5 text-primary-foreground" />
              </div>
              <span className="font-heading font-bold text-2xl gradient-text">Nexus</span>
            </Link>
          </div>

          <h1 className="font-heading font-bold text-2xl text-center text-foreground mb-1">
            Create an account
          </h1>
          <p className="text-sm text-muted-foreground text-center mb-8">
            Join thousands of shoppers on Nexus
          </p>

          {/* Google OAuth */}
          <Button
            id="google-register-btn"
            type="button"
            variant="outline"
            onClick={handleGoogleLogin}
            className="w-full h-11 border-border/60 hover:bg-secondary gap-3 mb-6 font-medium"
          >
            <svg className="w-5 h-5" viewBox="0 0 24 24">
              <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
              <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
              <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
              <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
            </svg>
            Sign up with Google
          </Button>

          <div className="relative flex items-center mb-6">
            <div className="flex-1 h-px bg-border/50" />
            <span className="px-4 text-xs text-muted-foreground">or register with email</span>
            <div className="flex-1 h-px bg-border/50" />
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Role selector */}
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-foreground">I am a</label>
              <div className="grid grid-cols-2 gap-2">
                {(["buyer", "seller"] as Role[]).map((r) => (
                  <button
                    key={r}
                    type="button"
                    onClick={() => setRole(r)}
                    className={`px-4 py-2.5 rounded-xl text-sm font-medium border transition-all duration-200 capitalize ${
                      role === r
                        ? "bg-primary/15 text-primary border-primary/40"
                        : "border-border/50 text-muted-foreground hover:text-foreground hover:border-border"
                    }`}
                  >
                    {r === "buyer" ? "🛍️ Buyer" : "🏪 Seller"}
                  </button>
                ))}
              </div>
            </div>

            {/* Name */}
            <div className="space-y-1.5">
              <label htmlFor="name" className="text-sm font-medium text-foreground">Full Name</label>
              <div className="relative">
                <User className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  id="name"
                  type="text"
                  placeholder="Pranay Kamble"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  required
                  autoComplete="name"
                  className="pl-10 h-11 bg-secondary border-border/50 focus:border-primary/50"
                />
              </div>
            </div>

            {/* Email */}
            <div className="space-y-1.5">
              <label htmlFor="email" className="text-sm font-medium text-foreground">Email</label>
              <div className="relative">
                <Mail className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  id="email"
                  type="email"
                  placeholder="you@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  autoComplete="email"
                  className="pl-10 h-11 bg-secondary border-border/50 focus:border-primary/50"
                />
              </div>
            </div>

            {/* Password */}
            <div className="space-y-1.5">
              <label htmlFor="password" className="text-sm font-medium text-foreground">Password</label>
              <div className="relative">
                <Lock className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  placeholder="Min. 8 characters"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  minLength={8}
                  autoComplete="new-password"
                  className="pl-10 pr-10 h-11 bg-secondary border-border/50 focus:border-primary/50"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                  {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
              </div>

              {/* Strength bar */}
              {password.length > 0 && (
                <div className="space-y-2 pt-1">
                  <div className="flex gap-1">
                    {[1, 2, 3].map((i) => (
                      <div
                        key={i}
                        className={`h-1 flex-1 rounded-full transition-all duration-300 ${
                          i <= passwordStrength ? strengthColors[passwordStrength] : "bg-border/50"
                        }`}
                      />
                    ))}
                  </div>
                  <p className={`text-xs font-medium ${passwordStrength === 3 ? "text-green-500" : passwordStrength === 2 ? "text-yellow-500" : "text-destructive"}`}>
                    {strengthLabels[passwordStrength]}
                  </p>
                  <div className="space-y-1">
                    {[
                      { label: "8+ characters", ok: passwordChecks.length },
                      { label: "Uppercase letter", ok: passwordChecks.upper },
                      { label: "Number", ok: passwordChecks.number },
                    ].map((c) => (
                      <div key={c.label} className="flex items-center gap-2">
                        <CheckCircle2 className={`w-3.5 h-3.5 ${c.ok ? "text-green-500" : "text-border"}`} />
                        <span className={`text-xs ${c.ok ? "text-foreground" : "text-muted-foreground"}`}>{c.label}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>

            <Button
              id="register-submit-btn"
              type="submit"
              disabled={loading}
              className="w-full h-11 bg-primary hover:bg-primary/90 glow-primary gap-2 font-semibold"
            >
              {loading ? (
                <><Loader2 className="w-4 h-4 animate-spin" />Creating account...</>
              ) : (
                <>Create Account<ArrowRight className="w-4 h-4" /></>
              )}
            </Button>
          </form>

          <p className="mt-6 text-center text-sm text-muted-foreground">
            Already have an account?{" "}
            <Link href="/auth/login" className="text-primary font-medium hover:underline">
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
