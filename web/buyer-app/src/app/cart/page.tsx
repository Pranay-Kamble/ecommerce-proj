"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import Image from "next/image";
import {
  ShoppingCart, Trash2, Plus, Minus, ArrowRight, Package, Loader2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { orderApi } from "@/lib/api";
import { useAuthStore, useCartStore } from "@/lib/store";
import toast from "react-hot-toast";

interface BackendCartItem {
  product_variant_id: string;
  quantity: number;
  price?: number;
  title?: string;
  image?: string;
}

export default function CartPage() {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();
  const { items, setItems, removeItem, clearCart } = useCartStore();

  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState<string | null>(null);

  useEffect(() => {
    if (!isAuthenticated) {
      router.replace("/auth/login");
      return;
    }
    fetchCart();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAuthenticated]);

  const fetchCart = async () => {
    setLoading(true);
    try {
      const res = await orderApi.get("/cart");
      const cartItems: BackendCartItem[] = res.data?.items ?? [];
      setItems(
        cartItems.map((i) => ({
          product_variant_id: i.product_variant_id,
          quantity: i.quantity,
          price: i.price,
          title: i.title,
        }))
      );
    } catch {
      // If cart fetch fails (backend down) keep local Zustand state
    } finally {
      setLoading(false);
    }
  };

  const handleRemove = async (variantId: string) => {
    setUpdating(variantId);
    try {
      await orderApi.delete(`/cart/remove/${variantId}`);
      removeItem(variantId);
      toast.success("Item removed");
    } catch {
      toast.error("Failed to remove item");
    } finally {
      setUpdating(null);
    }
  };

  const handleQuantityChange = async (variantId: string, delta: number, currentQty: number) => {
    const newQty = currentQty + delta;
    if (newQty <= 0) {
      await handleRemove(variantId);
      return;
    }
    setUpdating(variantId);
    try {
      // Backend add merges quantity — remove first then re-add for set semantics
      await orderApi.delete(`/cart/remove/${variantId}`);
      const res = await orderApi.post("/cart/add", {
        product_variant_id: variantId,
        quantity: newQty,
      });
      const cartItems: BackendCartItem[] = res.data?.items ?? [];
      setItems(
        cartItems.map((i) => ({
          product_variant_id: i.product_variant_id,
          quantity: i.quantity,
          price: i.price,
          title: i.title,
        }))
      );
    } catch {
      toast.error("Failed to update quantity");
    } finally {
      setUpdating(null);
    }
  };

  const handleClearCart = async () => {
    try {
      await orderApi.delete("/cart");
      clearCart();
      toast.success("Cart cleared");
    } catch {
      toast.error("Failed to clear cart");
    }
  };

  const subtotal = items.reduce(
    (sum, item) => sum + (item.price ?? 0) * item.quantity,
    0
  );

  const hasKnownPrices = items.some((i) => i.price);

  if (loading) {
    return (
      <div className="min-h-screen max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <Skeleton className="h-10 w-48 mb-8" />
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2 space-y-4">
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-28 w-full rounded-2xl" />
            ))}
          </div>
          <Skeleton className="h-64 rounded-2xl" />
        </div>
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center px-4 text-center">
        <div className="w-24 h-24 rounded-3xl bg-secondary flex items-center justify-center mb-6">
          <ShoppingCart className="w-12 h-12 text-muted-foreground" />
        </div>
        <h1 className="font-heading font-bold text-3xl text-foreground mb-3">Your cart is empty</h1>
        <p className="text-muted-foreground mb-8 max-w-sm">
          Looks like you haven&apos;t added anything yet. Browse our products and find something you love!
        </p>
        <Link href="/products">
          <Button className="bg-primary hover:bg-primary/90 gap-2">
            Browse Products <ArrowRight className="w-4 h-4" />
          </Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="min-h-screen max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <h1 className="font-heading font-bold text-3xl text-foreground">
          Shopping Cart
          <span className="ml-3 text-lg text-muted-foreground font-normal">
            ({items.length} item{items.length !== 1 ? "s" : ""})
          </span>
        </h1>
        <button
          onClick={handleClearCart}
          className="text-sm text-muted-foreground hover:text-destructive transition-colors flex items-center gap-1.5"
        >
          <Trash2 className="w-3.5 h-3.5" />
          Clear all
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Cart Items */}
        <div className="lg:col-span-2 space-y-4">
          {items.map((item) => (
            <div
              key={item.product_variant_id}
              className="glass rounded-2xl p-5 border border-border/50 flex gap-4 items-center card-hover"
            >
              {/* Placeholder image */}
              <div className="w-20 h-20 rounded-xl bg-secondary flex items-center justify-center shrink-0">
                <Package className="w-8 h-8 text-muted-foreground" />
              </div>

              <div className="flex-1 min-w-0">
                <p className="font-medium text-foreground text-sm truncate">
                  {item.title || "Product Variant"}
                </p>
                <p className="text-xs text-muted-foreground mt-0.5 font-mono truncate">
                  {item.product_variant_id}
                </p>
                {item.price && (
                  <p className="text-primary font-semibold mt-1">
                    ₹{item.price.toLocaleString("en-IN")}
                  </p>
                )}
              </div>

              {/* Qty controls */}
              <div className="flex items-center gap-2 shrink-0">
                <button
                  onClick={() => handleQuantityChange(item.product_variant_id, -1, item.quantity)}
                  disabled={updating === item.product_variant_id}
                  className="w-8 h-8 rounded-lg bg-secondary hover:bg-muted flex items-center justify-center transition-all disabled:opacity-50"
                >
                  {updating === item.product_variant_id ? (
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  ) : (
                    <Minus className="w-3.5 h-3.5" />
                  )}
                </button>
                <span className="w-8 text-center font-semibold text-sm">{item.quantity}</span>
                <button
                  onClick={() => handleQuantityChange(item.product_variant_id, 1, item.quantity)}
                  disabled={updating === item.product_variant_id}
                  className="w-8 h-8 rounded-lg bg-secondary hover:bg-muted flex items-center justify-center transition-all disabled:opacity-50"
                >
                  <Plus className="w-3.5 h-3.5" />
                </button>
              </div>

              {/* Line total */}
              {item.price && (
                <div className="text-right shrink-0 w-20">
                  <p className="font-bold text-foreground">
                    ₹{(item.price * item.quantity).toLocaleString("en-IN")}
                  </p>
                </div>
              )}

              <button
                onClick={() => handleRemove(item.product_variant_id)}
                disabled={updating === item.product_variant_id}
                className="p-2 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-all"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            </div>
          ))}
        </div>

        {/* Order Summary */}
        <div className="lg:col-span-1">
          <div className="glass rounded-2xl p-6 border border-border/50 sticky top-24">
            <h2 className="font-heading font-semibold text-lg text-foreground mb-5">Order Summary</h2>

            <div className="space-y-3 mb-5">
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Items ({items.length})</span>
                {hasKnownPrices ? (
                  <span className="text-foreground font-medium">₹{subtotal.toLocaleString("en-IN")}</span>
                ) : (
                  <span className="text-muted-foreground text-xs italic">Prices at checkout</span>
                )}
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Shipping</span>
                <span className="text-green-500 font-medium">Free</span>
              </div>
              {hasKnownPrices && (
                <>
                  <div className="h-px bg-border/50 my-2" />
                  <div className="flex justify-between font-semibold text-base">
                    <span className="text-foreground">Total</span>
                    <span className="text-primary">₹{subtotal.toLocaleString("en-IN")}</span>
                  </div>
                </>
              )}
            </div>

            <Button
              id="proceed-to-checkout-btn"
              onClick={() => router.push("/checkout")}
              className="w-full h-11 bg-primary hover:bg-primary/90 glow-primary gap-2 font-semibold"
            >
              Proceed to Checkout <ArrowRight className="w-4 h-4" />
            </Button>

            <Link href="/products" className="block mt-4 text-center text-sm text-muted-foreground hover:text-foreground transition-colors">
              ← Continue shopping
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
