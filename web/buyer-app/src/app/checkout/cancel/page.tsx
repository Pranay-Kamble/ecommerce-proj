"use client";

import Link from "next/link";
import { XCircle, ArrowLeft, ShoppingCart } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function CheckoutCancelPage() {
  return (
    <div className="min-h-screen hero-gradient flex items-center justify-center px-4">
      <div className="absolute top-1/3 right-1/3 w-72 h-72 bg-destructive/5 rounded-full blur-3xl pointer-events-none" />

      <div className="relative w-full max-w-md text-center">
        <div className="glass-strong rounded-3xl p-10 border border-border/60 shadow-2xl shadow-black/30">
          <div className="flex justify-center mb-6">
            <div className="w-20 h-20 rounded-full bg-destructive/10 border-2 border-destructive/30 flex items-center justify-center">
              <XCircle className="w-10 h-10 text-destructive" />
            </div>
          </div>

          <h1 className="font-heading font-bold text-3xl text-foreground mb-3">
            Payment Cancelled
          </h1>
          <p className="text-muted-foreground mb-8 leading-relaxed">
            Your payment was cancelled. No charge has been made. Your cart items are still saved — you can try again whenever you&apos;re ready.
          </p>

          <div className="space-y-3">
            <Link href="/cart" className="block">
              <Button className="w-full h-11 bg-primary hover:bg-primary/90 gap-2">
                <ShoppingCart className="w-4 h-4" />
                Return to Cart
              </Button>
            </Link>
            <Link href="/products" className="block">
              <Button variant="outline" className="w-full h-11 border-border/50 gap-2">
                <ArrowLeft className="w-4 h-4" />
                Continue Shopping
              </Button>
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
