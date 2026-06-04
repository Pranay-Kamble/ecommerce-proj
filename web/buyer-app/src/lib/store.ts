import { create } from "zustand";
import { persist } from "zustand/middleware";
import { CartItem } from "@/lib/types";

interface AuthState {
  token: string | null;
  userId: string | null;
  isAuthenticated: boolean;
  setToken: (token: string, userId: string) => void;
  logout: () => void;
  refreshAccessToken: () => Promise<boolean>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      userId: null,
      isAuthenticated: false,
      setToken: (token, userId) => {
        if (typeof window !== "undefined") {
          localStorage.setItem("access_token", token);
        }
        set({ token, userId, isAuthenticated: true });
      },
      logout: () => {
        if (typeof window !== "undefined") {
          localStorage.removeItem("access_token");
        }
        set({ token: null, userId: null, isAuthenticated: false });
      },
      // Silently renews the access token using the HttpOnly refreshToken cookie
      refreshAccessToken: async () => {
        try {
          const AUTH_BASE = process.env.NEXT_PUBLIC_AUTH_API_URL || "http://localhost:8080";
          const res = await fetch(`${AUTH_BASE}/api/v1/auth/refresh`, {
            method: "POST",
            credentials: "include", // sends the HttpOnly refreshToken cookie
          });
          if (!res.ok) {
            set({ token: null, userId: null, isAuthenticated: false });
            if (typeof window !== "undefined") localStorage.removeItem("access_token");
            return false;
          }
          const data = await res.json();
          const jwt: string = data.jwt;
          let userId = "";
          try {
            const payload = JSON.parse(atob(jwt.split(".")[1]));
            userId = payload.sub || payload.user_id || payload.id || "";
          } catch {}
          if (typeof window !== "undefined") localStorage.setItem("access_token", jwt);
          set({ token: jwt, userId, isAuthenticated: true });
          return true;
        } catch {
          return false;
        }
      },
    }),
    { name: "auth-store" }
  )
);

interface CartStore {
  items: CartItem[];
  itemCount: number;
  addItem: (item: CartItem) => void;
  removeItem: (variantId: string) => void;
  clearCart: () => void;
  setItems: (items: CartItem[]) => void;
}

export const useCartStore = create<CartStore>()(
  persist(
    (set, get) => ({
      items: [],
      itemCount: 0,
      addItem: (item) => {
        const items = get().items;
        const existing = items.find(
          (i) => i.product_variant_id === item.product_variant_id
        );
        let newItems: CartItem[];
        if (existing) {
          newItems = items.map((i) =>
            i.product_variant_id === item.product_variant_id
              ? { ...i, quantity: i.quantity + item.quantity }
              : i
          );
        } else {
          newItems = [...items, item];
        }
        set({ items: newItems, itemCount: newItems.reduce((s, i) => s + i.quantity, 0) });
      },
      removeItem: (variantId) => {
        const newItems = get().items.filter(
          (i) => i.product_variant_id !== variantId
        );
        set({ items: newItems, itemCount: newItems.reduce((s, i) => s + i.quantity, 0) });
      },
      clearCart: () => set({ items: [], itemCount: 0 }),
      setItems: (items) =>
        set({ items, itemCount: items.reduce((s, i) => s + i.quantity, 0) }),
    }),
    { name: "cart-store" }
  )
);
