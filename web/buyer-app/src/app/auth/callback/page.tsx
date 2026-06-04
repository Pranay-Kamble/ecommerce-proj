"use client";

import { useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Loader2, Zap } from "lucide-react";
import { useAuthStore } from "@/lib/store";
import toast from "react-hot-toast";
import { Suspense } from "react";

/**
 * This page handles the redirect from the backend after Google OAuth.
 * The backend (GET /google/callback) issues a JWT in the JSON body,
 * but since it's a server-side redirect, we need a frontend landing page.
 *
 * The backend redirects to: /auth/callback?jwt=<token>&msg=<message>
 * We capture the JWT from the query string, store it, and redirect to home.
 *
 * NOTE: If the backend does a full redirect (not to this page), the
 * Google flow works server-side and the user lands at / with the cookie set.
 * This page handles the case where the backend redirects here with the JWT.
 */
function OAuthCallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { setToken } = useAuthStore();

  useEffect(() => {
    const jwt = searchParams.get("jwt");
    const error = searchParams.get("error");

    if (error) {
      toast.error(`Auth failed: ${error}`);
      router.replace("/auth/login");
      return;
    }

    if (jwt) {
      let userId = "";
      try {
        const payload = JSON.parse(atob(jwt.split(".")[1]));
        userId = payload.sub || payload.user_id || payload.id || "";
      } catch {}

      setToken(jwt, userId);
      toast.success("Signed in with Google!");
      router.replace("/");
    } else {
      // No JWT in query — the backend may have set the cookie-based session
      // and the JWT is delivered via the backend response body.
      // For now redirect home; the refresh endpoint will handle token renewal.
      router.replace("/");
    }
  }, [searchParams, setToken, router]);

  return (
    <div className="min-h-screen hero-gradient flex items-center justify-center">
      <div className="text-center space-y-6">
        <div className="flex justify-center">
          <div className="w-16 h-16 rounded-2xl bg-primary/15 border border-primary/30 flex items-center justify-center">
            <Zap className="w-8 h-8 text-primary" />
          </div>
        </div>
        <div className="space-y-2">
          <div className="flex items-center justify-center gap-3">
            <Loader2 className="w-5 h-5 animate-spin text-primary" />
            <p className="text-foreground font-medium">Completing sign in...</p>
          </div>
          <p className="text-sm text-muted-foreground">Please wait while we set up your session</p>
        </div>
      </div>
    </div>
  );
}

export default function OAuthCallbackPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen hero-gradient flex items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    }>
      <OAuthCallbackContent />
    </Suspense>
  );
}
