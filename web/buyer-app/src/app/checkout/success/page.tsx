"use client";

import { Suspense } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { CheckCircle2, Package, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useCartStore } from "@/lib/store";
import { useEffect } from "react";

function SuccessContent() {
  const searchParams = useSearchParams();
  const orderId = searchParams.get("order_id") ?? searchParams.get("session_id") ?? "";
  const { clearCart } = useCartStore();

  // Clear local cart on success (backend already cleared the Redis cart)
  useEffect(() => {
    clearCart();
  }, [clearCart]);

  return (
    <div className="min-h-screen hero-gradient flex items-center justify-center px-4">
      <div className="absolute top-1/3 left-1/3 w-72 h-72 bg-green-500/5 rounded-full blur-3xl pointer-events-none" />

      <div className="relative w-full max-w-md text-center">
        <div className="glass-strong rounded-3xl p-10 border border-border/60 shadow-2xl shadow-black/30">
          {/* Success icon */}
          <div className="flex justify-center mb-6">
            <div className="w-20 h-20 rounded-full bg-green-500/15 border-2 border-green-500/40 flex items-center justify-center">
              <CheckCircle2 className="w-10 h-10 text-green-500" />
            </div>
          </div>

          <h1 className="font-heading font-bold text-3xl text-foreground mb-3">
            Payment Successful!
          </h1>
          <p className="text-muted-foreground mb-6 leading-relaxed">
            Your order has been confirmed and is being processed. You&apos;ll receive an email once it ships.
          </p>

          {orderId && (
            <div className="bg-secondary rounded-xl px-4 py-3 mb-6 border border-border/50">
              <p className="text-xs text-muted-foreground mb-1">Order ID</p>
              <p className="font-mono font-semibold text-primary text-sm">{orderId}</p>
            </div>
          )}

          <div className="space-y-3">
            {orderId && (
              <Link href={`/orders/${orderId}`} className="block">
                <Button className="w-full h-11 bg-primary hover:bg-primary/90 gap-2">
                  <Package className="w-4 h-4" />
                  View Order Details
                </Button>
              </Link>
            )}
            <Link href="/orders" className="block">
              <Button variant="outline" className="w-full h-11 border-border/50 gap-2">
                My Orders
              </Button>
            </Link>
            <Link href="/products">
              <p className="text-sm text-muted-foreground hover:text-foreground transition-colors mt-2 flex items-center justify-center gap-1">
                Continue Shopping <ArrowRight className="w-3.5 h-3.5" />
              </p>
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function CheckoutSuccessPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen hero-gradient flex items-center justify-center">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    }>
      <SuccessContent />
    </Suspense>
  );
}
