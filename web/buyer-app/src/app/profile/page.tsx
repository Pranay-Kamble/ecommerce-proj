"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { User, Phone, MapPin, Plus, Loader2, Save, CheckCircle2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { orderApi } from "@/lib/api";
import { useAuthStore } from "@/lib/store";
import toast from "react-hot-toast";

interface Address {
  title?: string;
  address_line: string;
  city: string;
  state: string;
  zip_code: string;
  is_default?: boolean;
}

interface CustomerProfile {
  name: string;
  phone: string;
  addresses?: Address[];
}

export default function ProfilePage() {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();

  const [profile, setProfile] = useState<CustomerProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [addingAddress, setAddingAddress] = useState(false);
  const [showAddForm, setShowAddForm] = useState(false);

  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");

  // New address form
  const [addrTitle, setAddrTitle] = useState("");
  const [addrLine, setAddrLine] = useState("");
  const [addrCity, setAddrCity] = useState("");
  const [addrState, setAddrState] = useState("");
  const [addrZip, setAddrZip] = useState("");
  const [addrDefault, setAddrDefault] = useState(false);

  useEffect(() => {
    if (!isAuthenticated) { router.replace("/auth/login"); return; }
    fetchProfile();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAuthenticated]);

  const fetchProfile = async () => {
    setLoading(true);
    try {
      const res = await orderApi.get<CustomerProfile>("/profile");
      setProfile(res.data);
      setName(res.data.name ?? "");
      setPhone(res.data.phone ?? "");
    } catch {
      setProfile(null);
    } finally {
      setLoading(false);
    }
  };

  const handleSaveProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      const res = await orderApi.post("/profile", { name, phone });
      setProfile(res.data);
      toast.success("Profile updated!");
    } catch (err: any) {
      toast.error(err?.response?.data?.error ?? "Failed to save profile");
    } finally {
      setSaving(false);
    }
  };

  const handleAddAddress = async (e: React.FormEvent) => {
    e.preventDefault();
    setAddingAddress(true);
    try {
      await orderApi.post("/profile/addresses", {
        title: addrTitle,
        address_line: addrLine,
        city: addrCity,
        state: addrState,
        zip_code: addrZip,
        is_default: addrDefault,
      });
      toast.success("Address saved!");
      setShowAddForm(false);
      setAddrTitle(""); setAddrLine(""); setAddrCity(""); setAddrState(""); setAddrZip("");
      fetchProfile();
    } catch (err: any) {
      toast.error(err?.response?.data?.error ?? "Failed to save address");
    } finally {
      setAddingAddress(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen max-w-2xl mx-auto px-4 py-12">
        <Skeleton className="h-10 w-48 mb-8" />
        <Skeleton className="h-48 rounded-2xl mb-4" />
        <Skeleton className="h-64 rounded-2xl" />
      </div>
    );
  }

  return (
    <div className="min-h-screen max-w-2xl mx-auto px-4 sm:px-6 py-12">
      <h1 className="font-heading font-bold text-3xl text-foreground mb-8">My Profile</h1>

      {/* Profile info */}
      <div className="glass rounded-2xl border border-border/50 p-6 mb-6">
        <h2 className="font-heading font-semibold text-foreground mb-5 flex items-center gap-2">
          <User className="w-4 h-4 text-primary" /> Personal Information
        </h2>
        <form onSubmit={handleSaveProfile} className="space-y-4">
          <div className="space-y-1.5">
            <label htmlFor="profile-name" className="text-sm font-medium text-foreground">Full Name</label>
            <div className="relative">
              <User className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input id="profile-name" value={name} onChange={(e) => setName(e.target.value)} required
                placeholder="Pranay Kamble" className="pl-10 h-11 bg-secondary border-border/50 focus:border-primary/50" />
            </div>
          </div>
          <div className="space-y-1.5">
            <label htmlFor="profile-phone" className="text-sm font-medium text-foreground">Phone Number</label>
            <div className="relative">
              <Phone className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input id="profile-phone" value={phone} onChange={(e) => setPhone(e.target.value)}
                placeholder="+91 9876543210" type="tel" className="pl-10 h-11 bg-secondary border-border/50 focus:border-primary/50" />
            </div>
          </div>
          <Button type="submit" disabled={saving} className="bg-primary hover:bg-primary/90 gap-2">
            {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
            Save Changes
          </Button>
        </form>
      </div>

      {/* Saved Addresses */}
      <div className="glass rounded-2xl border border-border/50 p-6">
        <div className="flex items-center justify-between mb-5">
          <h2 className="font-heading font-semibold text-foreground flex items-center gap-2">
            <MapPin className="w-4 h-4 text-primary" /> Saved Addresses
          </h2>
          <Button variant="outline" size="sm" onClick={() => setShowAddForm(!showAddForm)} className="border-border/50 gap-1.5 text-xs">
            <Plus className="w-3.5 h-3.5" /> Add Address
          </Button>
        </div>

        {/* Existing addresses */}
        {profile?.addresses?.length ? (
          <div className="space-y-3 mb-4">
            {profile.addresses.map((addr, idx) => (
              <div key={idx} className="rounded-xl bg-secondary border border-border/50 p-4 text-sm">
                <div className="flex items-center gap-2 mb-1">
                  {addr.title && <span className="font-medium text-foreground">{addr.title}</span>}
                  {addr.is_default && (
                    <span className="text-xs px-2 py-0.5 bg-primary/15 text-primary rounded-full border border-primary/30">Default</span>
                  )}
                </div>
                <p className="text-muted-foreground leading-relaxed">
                  {addr.address_line}, {addr.city}, {addr.state} — {addr.zip_code}
                </p>
              </div>
            ))}
          </div>
        ) : !showAddForm ? (
          <p className="text-sm text-muted-foreground py-4 text-center">No saved addresses yet.</p>
        ) : null}

        {/* Add address form */}
        {showAddForm && (
          <form onSubmit={handleAddAddress} className="space-y-3 border-t border-border/40 pt-5 mt-4">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-foreground">Title (optional)</label>
                <Input value={addrTitle} onChange={(e) => setAddrTitle(e.target.value)}
                  placeholder="Home / Office" className="h-9 text-sm bg-secondary border-border/50" />
              </div>
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-foreground">PIN Code *</label>
                <Input value={addrZip} onChange={(e) => setAddrZip(e.target.value)} required
                  placeholder="411001" className="h-9 text-sm bg-secondary border-border/50" />
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-foreground">Address Line *</label>
              <Input value={addrLine} onChange={(e) => setAddrLine(e.target.value)} required
                placeholder="123 MG Road, Flat 4B" className="h-9 text-sm bg-secondary border-border/50" />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-foreground">City *</label>
                <Input value={addrCity} onChange={(e) => setAddrCity(e.target.value)} required
                  placeholder="Pune" className="h-9 text-sm bg-secondary border-border/50" />
              </div>
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-foreground">State *</label>
                <Input value={addrState} onChange={(e) => setAddrState(e.target.value)} required
                  placeholder="Maharashtra" className="h-9 text-sm bg-secondary border-border/50" />
              </div>
            </div>
            <label className="flex items-center gap-2 cursor-pointer text-sm text-foreground">
              <input type="checkbox" checked={addrDefault} onChange={(e) => setAddrDefault(e.target.checked)}
                className="rounded border-border/50 accent-primary" />
              Set as default address
            </label>
            <div className="flex gap-2 pt-1">
              <Button type="submit" disabled={addingAddress} className="bg-primary hover:bg-primary/90 gap-2 text-sm h-9">
                {addingAddress ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <CheckCircle2 className="w-3.5 h-3.5" />}
                Save Address
              </Button>
              <Button type="button" variant="outline" onClick={() => setShowAddForm(false)} className="border-border/50 text-sm h-9">
                Cancel
              </Button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
