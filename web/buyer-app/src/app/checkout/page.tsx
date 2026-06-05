"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { MapPin, Phone, User, Loader2, ArrowRight, ShoppingBag, CreditCard } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { orderApi } from "@/lib/api";
import { useAuthStore, useCartStore } from "@/lib/store";
import toast from "react-hot-toast";

interface ProfileData {
  name: string;
  phone: string;
  addresses?: Array<{
    title?: string;
    address_line: string;
    city: string;
    state: string;
    zip_code: string;
    is_default?: boolean;
  }>;
}

export default function CheckoutPage() {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();
  const { items } = useCartStore();

  const [loading, setLoading] = useState(false);
  const [prefilling, setPrefilling] = useState(true);

  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [addressLine, setAddressLine] = useState("");
  const [city, setCity] = useState("");
  const [state, setState] = useState("");
  const [zipCode, setZipCode] = useState("");

  useEffect(() => {
    if (!isAuthenticated) {
      router.replace("/auth/login");
      return;
    }
    if (items.length === 0) {
      router.replace("/cart");
      return;
    }
    prefillFromProfile();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAuthenticated, items.length]);

  const prefillFromProfile = async () => {
    setPrefilling(true);
    try {
      const res = await orderApi.get<ProfileData>("/profile");
      const profile = res.data;
      if (profile.name) setName(profile.name);
      if (profile.phone) setPhone(profile.phone);
      const defaultAddr = profile.addresses?.find((a) => a.is_default) ?? profile.addresses?.[0];
      if (defaultAddr) {
        setAddressLine(defaultAddr.address_line);
        setCity(defaultAddr.city);
        setState(defaultAddr.state);
        setZipCode(defaultAddr.zip_code);
      }
    } catch {
      // No profile yet — user fills it in manually
    } finally {
      setPrefilling(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const res = await orderApi.post("/checkout", {
        name,
        phone,
        address_line: addressLine,
        city,
        state,
        zip_code: zipCode,
      });
      const { payment_url, order_id } = res.data;
      toast.success("Order created! Redirecting to payment...");
      // Redirect to Stripe hosted checkout
      window.location.href = payment_url;
    } catch (err: any) {
      const msg = err?.response?.data?.error || "Checkout failed. Please try again.";
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  };

  const subtotal = items.reduce((s, i) => s + (i.price ?? 0) * i.quantity, 0);

  const InputField = ({
    id, label, icon: Icon, value, onChange, placeholder, type = "text", required = true,
  }: {
    id: string; label: string; icon: React.ElementType; value: string;
    onChange: (v: string) => void; placeholder: string; type?: string; required?: boolean;
  }) => (
    <div className="space-y-1.5">
      <label htmlFor={id} className="text-sm font-medium text-foreground">{label}</label>
      <div className="relative">
        <Icon className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
        <Input
          id={id}
          type={type}
          placeholder={placeholder}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          required={required}
          className="pl-10 h-11 bg-secondary border-border/50 focus:border-primary/50"
        />
      </div>
    </div>
  );

  return (
    <div className="min-h-screen max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <h1 className="font-heading font-bold text-3xl text-foreground mb-8">Checkout</h1>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Shipping Form */}
        <div className="lg:col-span-2">
          <div className="glass rounded-2xl p-6 border border-border/50">
            <h2 className="font-heading font-semibold text-lg text-foreground mb-5 flex items-center gap-2">
              <MapPin className="w-5 h-5 text-primary" />
              Shipping Details
            </h2>

            {prefilling ? (
              <div className="flex items-center gap-2 text-sm text-muted-foreground py-4">
                <Loader2 className="w-4 h-4 animate-spin" />
                Loading your saved details...
              </div>
            ) : (
              <form id="checkout-form" onSubmit={handleSubmit} className="space-y-4">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <InputField id="checkout-name" label="Full Name" icon={User} value={name} onChange={setName} placeholder="Pranay Kamble" />
                  <InputField id="checkout-phone" label="Phone Number" icon={Phone} value={phone} onChange={setPhone} placeholder="+91 9876543210" type="tel" />
                </div>
                <InputField id="checkout-address" label="Address Line" icon={MapPin} value={addressLine} onChange={setAddressLine} placeholder="123 MG Road, Flat 4B" />
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <InputField id="checkout-city" label="City" icon={MapPin} value={city} onChange={setCity} placeholder="Pune" />
                  <InputField id="checkout-state" label="State" icon={MapPin} value={state} onChange={setState} placeholder="Maharashtra" />
                  <InputField id="checkout-zip" label="PIN Code" icon={MapPin} value={zipCode} onChange={setZipCode} placeholder="411001" />
                </div>
              </form>
            )}
          </div>
        </div>

        {/* Order Summary */}
        <div className="lg:col-span-1">
          <div className="glass rounded-2xl p-6 border border-border/50 sticky top-24 space-y-5">
            <h2 className="font-heading font-semibold text-lg text-foreground flex items-center gap-2">
              <ShoppingBag className="w-5 h-5 text-primary" />
              Your Order
            </h2>

            <div className="space-y-3 max-h-48 overflow-y-auto pr-1">
              {items.map((item) => (
                <div key={item.product_variant_id} className="flex justify-between text-sm">
                  <span className="text-muted-foreground truncate max-w-[60%]">
                    {item.title || item.product_variant_id.slice(0, 10) + "…"} × {item.quantity}
                  </span>
                  {item.price && (
                    <span className="text-foreground font-medium shrink-0">
                      ₹{(item.price * item.quantity).toLocaleString("en-IN")}
                    </span>
                  )}
                </div>
              ))}
            </div>

            <div className="h-px bg-border/50" />

            <div className="flex justify-between font-semibold">
              <span className="text-foreground">
                {subtotal > 0 ? "Total" : "Total (at checkout)"}
              </span>
              {subtotal > 0 ? (
                <span className="text-primary text-lg">₹{subtotal.toLocaleString("en-IN")}</span>
              ) : (
                <span className="text-muted-foreground text-sm italic">Calculated by backend</span>
              )}
            </div>

            <Button
              id="place-order-btn"
              type="submit"
              form="checkout-form"
              disabled={loading || prefilling}
              className="w-full h-11 bg-primary hover:bg-primary/90 glow-primary gap-2 font-semibold"
            >
              {loading ? (
                <><Loader2 className="w-4 h-4 animate-spin" />Processing...</>
              ) : (
                <><CreditCard className="w-4 h-4" />Pay with Stripe</>
              )}
            </Button>

            <p className="text-xs text-muted-foreground text-center">
              🔒 Secured by Stripe. You&apos;ll be redirected to complete payment.
            </p>

            <Link href="/cart" className="block text-center text-sm text-muted-foreground hover:text-foreground transition-colors">
              ← Back to cart
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
