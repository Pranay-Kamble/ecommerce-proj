"use client";

import { useState, useRef, useEffect } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Mail, Loader2, Zap, ArrowRight, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { authApi } from "@/lib/api";
import { useAuthStore } from "@/lib/store";
import toast from "react-hot-toast";
import { Suspense } from "react";

function VerifyContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { setToken } = useAuthStore();

  const prefillEmail = searchParams.get("email") ?? "";

  const [email, setEmail] = useState(prefillEmail);
  const [otp, setOtp] = useState<string[]>(["", "", "", "", "", ""]);
  const [loading, setLoading] = useState(false);
  const [resending, setResending] = useState(false);
  const [countdown, setCountdown] = useState(0);

  const inputRefs = useRef<Array<HTMLInputElement | null>>([]);

  // Countdown timer for resend
  useEffect(() => {
    if (countdown <= 0) return;
    const t = setTimeout(() => setCountdown((c) => c - 1), 1000);
    return () => clearTimeout(t);
  }, [countdown]);

  const handleOtpChange = (index: number, value: string) => {
    // Only allow single digit
    const digit = value.replace(/\D/g, "").slice(-1);
    const newOtp = [...otp];
    newOtp[index] = digit;
    setOtp(newOtp);

    // Auto-advance focus
    if (digit && index < 5) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleOtpKeyDown = (index: number, e: React.KeyboardEvent) => {
    if (e.key === "Backspace" && !otp[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
    // Allow paste
    if (e.key === "v" && (e.ctrlKey || e.metaKey)) return;
  };

  const handleOtpPaste = (e: React.ClipboardEvent) => {
    const paste = e.clipboardData.getData("text").replace(/\D/g, "").slice(0, 6);
    if (paste.length === 6) {
      setOtp(paste.split(""));
      inputRefs.current[5]?.focus();
      e.preventDefault();
    }
  };

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    const otpString = otp.join("");
    if (otpString.length !== 6) {
      toast.error("Please enter the full 6-digit OTP");
      return;
    }
    setLoading(true);

    try {
      const res = await authApi.post("/verify", { email, otp: otpString });
      const { jwt } = res.data;

      let userId = "";
      try {
        const payload = JSON.parse(atob(jwt.split(".")[1]));
        userId = payload.sub || payload.user_id || payload.id || "";
      } catch {}

      setToken(jwt, userId);
      toast.success("Email verified! Welcome to Nexus 🎉");
      router.push("/");
    } catch (err: any) {
      const msg = err?.response?.data?.error || "Verification failed";
      toast.error(msg);
      // Clear OTP on failure
      setOtp(["", "", "", "", "", ""]);
      inputRefs.current[0]?.focus();
    } finally {
      setLoading(false);
    }
  };

  const handleResend = async () => {
    if (!email) {
      toast.error("Please enter your email address");
      return;
    }
    setResending(true);
    try {
      await authApi.post("/resend-otp", { email });
      toast.success("New OTP sent to your email!");
      setCountdown(60);
    } catch (err: any) {
      const msg = err?.response?.data?.error || "Failed to resend OTP";
      toast.error(msg);
    } finally {
      setResending(false);
    }
  };

  return (
    <div className="min-h-screen hero-gradient flex items-center justify-center px-4 py-20">
      <div className="absolute top-1/3 left-1/3 w-72 h-72 bg-primary/5 rounded-full blur-3xl pointer-events-none" />

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

          {/* Icon */}
          <div className="flex justify-center mb-4">
            <div className="w-16 h-16 rounded-2xl bg-primary/15 border border-primary/30 flex items-center justify-center">
              <Mail className="w-8 h-8 text-primary" />
            </div>
          </div>

          <h1 className="font-heading font-bold text-2xl text-center text-foreground mb-1">
            Check your email
          </h1>
          <p className="text-sm text-muted-foreground text-center mb-8">
            We sent a 6-digit code to{" "}
            {prefillEmail ? (
              <span className="text-foreground font-medium">{prefillEmail}</span>
            ) : (
              "your email"
            )}
          </p>

          <form onSubmit={handleVerify} className="space-y-6">
            {/* Email field (if not pre-filled) */}
            {!prefillEmail && (
              <div className="space-y-1.5">
                <label htmlFor="verify-email" className="text-sm font-medium text-foreground">
                  Email
                </label>
                <div className="relative">
                  <Mail className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                  <Input
                    id="verify-email"
                    type="email"
                    placeholder="you@example.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                    className="pl-10 h-11 bg-secondary border-border/50 focus:border-primary/50"
                  />
                </div>
              </div>
            )}

            {/* OTP Input Grid */}
            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground">Verification Code</label>
              <div className="flex gap-2 justify-center" onPaste={handleOtpPaste}>
                {otp.map((digit, index) => (
                  <input
                    key={index}
                    ref={(el) => { inputRefs.current[index] = el; }}
                    id={`otp-${index}`}
                    type="text"
                    inputMode="numeric"
                    pattern="[0-9]*"
                    maxLength={1}
                    value={digit}
                    onChange={(e) => handleOtpChange(index, e.target.value)}
                    onKeyDown={(e) => handleOtpKeyDown(index, e)}
                    className={`w-12 h-14 text-center text-xl font-bold rounded-xl border-2 bg-secondary transition-all duration-200 focus:outline-none
                      ${digit
                        ? "border-primary text-primary"
                        : "border-border/50 text-foreground focus:border-primary/70"
                      }`}
                  />
                ))}
              </div>
            </div>

            <Button
              id="verify-submit-btn"
              type="submit"
              disabled={loading || otp.join("").length < 6}
              className="w-full h-11 bg-primary hover:bg-primary/90 glow-primary gap-2 font-semibold"
            >
              {loading ? (
                <><Loader2 className="w-4 h-4 animate-spin" />Verifying...</>
              ) : (
                <>Verify Email<ArrowRight className="w-4 h-4" /></>
              )}
            </Button>
          </form>

          {/* Resend */}
          <div className="mt-6 text-center">
            <p className="text-sm text-muted-foreground mb-3">
              Didn&apos;t receive the code?
            </p>
            {countdown > 0 ? (
              <p className="text-sm text-muted-foreground">
                Resend in <span className="text-primary font-medium">{countdown}s</span>
              </p>
            ) : (
              <button
                id="resend-otp-btn"
                onClick={handleResend}
                disabled={resending}
                className="flex items-center gap-2 mx-auto text-sm text-primary font-medium hover:underline disabled:opacity-50"
              >
                {resending ? (
                  <Loader2 className="w-3.5 h-3.5 animate-spin" />
                ) : (
                  <RefreshCw className="w-3.5 h-3.5" />
                )}
                Resend OTP
              </button>
            )}
          </div>

          <div className="mt-4 text-center">
            <Link href="/auth/login" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
              ← Back to login
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function VerifyPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen hero-gradient flex items-center justify-center">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    }>
      <VerifyContent />
    </Suspense>
  );
}
