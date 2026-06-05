"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft, Download, XCircle, MapPin, Phone, User,
  Package, Loader2, FileText,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { orderApi } from "@/lib/api";
import { useAuthStore } from "@/lib/store";
import { Order } from "@/lib/types";

const statusStyles: Record<string, { color: string; label: string }> = {
  pending:   { color: "bg-yellow-500/15 text-yellow-400 border-yellow-500/30", label: "Pending Payment" },
  paid:      { color: "bg-green-500/15 text-green-400 border-green-500/30",  label: "Paid" },
  cancelled: { color: "bg-destructive/15 text-destructive border-destructive/30", label: "Cancelled" },
  shipped:   { color: "bg-blue-500/15 text-blue-400 border-blue-500/30", label: "Shipped" },
  delivered: { color: "bg-primary/15 text-primary border-primary/30", label: "Delivered" },
};

export default function OrderDetailPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const { isAuthenticated } = useAuthStore();

  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(true);
  const [cancelling, setCancelling] = useState(false);
  const [downloadingInvoice, setDownloadingInvoice] = useState(false);

  useEffect(() => {
    if (!isAuthenticated) {
      router.replace("/auth/login");
      return;
    }
    fetchOrder();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAuthenticated, params.id]);

  const fetchOrder = async () => {
    setLoading(true);
    try {
      const res = await orderApi.get<Order>(`/orders/${params.id}`);
      setOrder(res.data);
    } catch {
      setOrder(null);
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = async () => {
    if (!confirm("Are you sure you want to cancel this order?")) return;
    setCancelling(true);
    try {
      await orderApi.patch(`/orders/${params.id}/cancel`);
      setOrder((prev) => prev ? { ...prev, status: "cancelled" } : prev);
    } catch (err: any) {
      alert(err?.response?.data?.error ?? "Failed to cancel order");
    } finally {
      setCancelling(false);
    }
  };

  const handleDownloadInvoice = async () => {
    setDownloadingInvoice(true);
    try {
      const res = await orderApi.get(`/orders/${params.id}/invoice`);
      const { download_url } = res.data;
      window.open(download_url, "_blank");
    } catch (err: any) {
      alert(err?.response?.data?.error ?? "Invoice not available yet");
    } finally {
      setDownloadingInvoice(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen max-w-3xl mx-auto px-4 py-12">
        <Skeleton className="h-6 w-32 mb-8" />
        <Skeleton className="h-48 rounded-2xl mb-4" />
        <Skeleton className="h-64 rounded-2xl" />
      </div>
    );
  }

  if (!order) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center px-4 text-center">
        <Package className="w-16 h-16 text-muted-foreground mb-4" />
        <h2 className="font-heading font-semibold text-2xl text-foreground mb-2">Order Not Found</h2>
        <p className="text-muted-foreground mb-6">This order doesn&apos;t exist or you don&apos;t have access.</p>
        <Link href="/orders"><Button variant="outline" className="border-border/50">← Back to Orders</Button></Link>
      </div>
    );
  }

  const statusStyle = statusStyles[order.status] ?? statusStyles.pending;

  return (
    <div className="min-h-screen max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      {/* Back */}
      <Link href="/orders" className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground mb-8 transition-colors w-fit">
        <ArrowLeft className="w-4 h-4" />
        Back to Orders
      </Link>

      {/* Header */}
      <div className="flex items-start justify-between mb-6 gap-4 flex-wrap">
        <div>
          <h1 className="font-heading font-bold text-2xl text-foreground mb-1">{order.id}</h1>
          <p className="text-sm text-muted-foreground">
            Placed on {new Date(order.created_at).toLocaleDateString("en-IN", {
              weekday: "long", day: "numeric", month: "long", year: "numeric",
            })}
          </p>
        </div>
        <span className={`text-sm font-medium px-3 py-1.5 rounded-full border ${statusStyle.color}`}>
          {statusStyle.label}
        </span>
      </div>

      {/* Actions */}
      <div className="flex gap-3 mb-6 flex-wrap">
        {order.status === "paid" && (
          <Button
            id="download-invoice-btn"
            onClick={handleDownloadInvoice}
            disabled={downloadingInvoice}
            variant="outline"
            className="border-border/50 gap-2"
          >
            {downloadingInvoice ? <Loader2 className="w-4 h-4 animate-spin" /> : <Download className="w-4 h-4" />}
            Download Invoice
          </Button>
        )}
        {order.status === "pending" && (
          <Button
            id="cancel-order-btn"
            onClick={handleCancel}
            disabled={cancelling}
            variant="outline"
            className="border-destructive/50 text-destructive hover:bg-destructive/10 gap-2"
          >
            {cancelling ? <Loader2 className="w-4 h-4 animate-spin" /> : <XCircle className="w-4 h-4" />}
            Cancel Order
          </Button>
        )}
      </div>

      {/* Order items */}
      <div className="glass rounded-2xl border border-border/50 p-6 mb-4">
        <h2 className="font-heading font-semibold text-foreground mb-4 flex items-center gap-2">
          <Package className="w-4 h-4 text-primary" /> Items
        </h2>
        <div className="space-y-3">
          {(order.items ?? []).map((item, idx) => (
            <div key={idx} className="flex justify-between items-center py-2 border-b border-border/30 last:border-0">
              <div>
                <p className="text-sm font-medium text-foreground font-mono">{item.product_id}</p>
                <p className="text-xs text-muted-foreground">Qty: {item.quantity}</p>
              </div>
              <p className="font-semibold text-foreground">₹{(item.price * item.quantity).toLocaleString("en-IN")}</p>
            </div>
          ))}
        </div>
        <div className="flex justify-between pt-4 mt-4 border-t border-border/50 font-semibold">
          <span className="text-foreground">Total</span>
          <span className="text-primary text-lg">₹{order.total_amount.toLocaleString("en-IN")}</span>
        </div>
      </div>

      {/* Shipping info */}
      <div className="glass rounded-2xl border border-border/50 p-6">
        <h2 className="font-heading font-semibold text-foreground mb-4 flex items-center gap-2">
          <MapPin className="w-4 h-4 text-primary" /> Shipping Address
        </h2>
        <div className="space-y-2 text-sm">
          <div className="flex items-center gap-2 text-foreground">
            <User className="w-4 h-4 text-muted-foreground shrink-0" />
            {order.shipping_name}
          </div>
          <div className="flex items-center gap-2 text-foreground">
            <Phone className="w-4 h-4 text-muted-foreground shrink-0" />
            {order.shipping_phone}
          </div>
          <div className="flex items-start gap-2 text-muted-foreground leading-relaxed">
            <MapPin className="w-4 h-4 mt-0.5 shrink-0" />
            <span>
              {order.shipping_address}, {order.shipping_city}, {order.shipping_state} — {order.shipping_zip}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
