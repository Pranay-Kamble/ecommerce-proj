"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Package, ChevronRight, ShoppingBag, ArrowRight, Loader2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { orderApi } from "@/lib/api";
import { useAuthStore } from "@/lib/store";
import { Order } from "@/lib/types";

const statusStyles: Record<string, { color: string; label: string }> = {
  pending:   { color: "bg-yellow-500/15 text-yellow-400 border-yellow-500/30", label: "Pending" },
  paid:      { color: "bg-green-500/15 text-green-400 border-green-500/30",  label: "Paid" },
  cancelled: { color: "bg-destructive/15 text-destructive border-destructive/30", label: "Cancelled" },
  shipped:   { color: "bg-blue-500/15 text-blue-400 border-blue-500/30", label: "Shipped" },
  delivered: { color: "bg-primary/15 text-primary border-primary/30", label: "Delivered" },
};

function StatusBadge({ status }: { status: string }) {
  const style = statusStyles[status] ?? statusStyles.pending;
  return (
    <span className={`text-xs font-medium px-2.5 py-1 rounded-full border ${style.color}`}>
      {style.label}
    </span>
  );
}

export default function OrdersPage() {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!isAuthenticated) {
      router.replace("/auth/login");
      return;
    }
    fetchOrders();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAuthenticated]);

  const fetchOrders = async () => {
    setLoading(true);
    try {
      const res = await orderApi.get<Order[]>("/orders");
      // Backend sends `id` as the public_id via JSON tag
      setOrders(res.data ?? []);
    } catch {
      setOrders([]);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <Skeleton className="h-10 w-48 mb-8" />
        <div className="space-y-4">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-24 rounded-2xl" />)}
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <h1 className="font-heading font-bold text-3xl text-foreground mb-8">My Orders</h1>

      {orders.length === 0 ? (
        <div className="text-center py-24 glass rounded-2xl border border-border/50">
          <ShoppingBag className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
          <h3 className="font-heading font-semibold text-xl text-foreground mb-2">No orders yet</h3>
          <p className="text-muted-foreground mb-8">Start shopping to see your orders here.</p>
          <Link href="/products">
            <Button className="bg-primary hover:bg-primary/90 gap-2">
              Browse Products <ArrowRight className="w-4 h-4" />
            </Button>
          </Link>
        </div>
      ) : (
        <div className="space-y-4">
          {orders.map((order) => (
            <Link key={order.id} href={`/orders/${order.id}`}>
              <div className="glass rounded-2xl p-5 border border-border/50 card-hover flex items-center gap-4 group cursor-pointer">
                <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center shrink-0">
                  <Package className="w-6 h-6 text-primary" />
                </div>

                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-3 mb-1">
                    <p className="font-mono font-semibold text-foreground text-sm">{order.id}</p>
                    <StatusBadge status={order.status} />
                  </div>
                  <p className="text-sm text-muted-foreground">
                    {order.items?.length ?? 0} item{(order.items?.length ?? 0) !== 1 ? "s" : ""} ·{" "}
                    {new Date(order.created_at).toLocaleDateString("en-IN", {
                      day: "numeric", month: "short", year: "numeric",
                    })}
                  </p>
                </div>

                <div className="text-right shrink-0">
                  <p className="font-semibold text-foreground">
                    ₹{order.total_amount.toLocaleString("en-IN")}
                  </p>
                </div>

                <ChevronRight className="w-4 h-4 text-muted-foreground group-hover:text-foreground transition-colors shrink-0" />
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
