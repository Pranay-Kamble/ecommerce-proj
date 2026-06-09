"use client";

import { useEffect, useState } from "react";
import { useParams, notFound } from "next/navigation";
import Image from "next/image";
import Link from "next/link";
import {
  ChevronLeft,
  ShoppingCart,
  Star,
  Package,
  CheckCircle2,
  XCircle,
  Loader2,
  Heart,
  Share2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { catalogApi, orderApi } from "@/lib/api";
import { Product, ProductVariant } from "@/lib/types";
import { useCartStore, useAuthStore } from "@/lib/store";
import toast from "react-hot-toast";

export default function ProductDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedVariant, setSelectedVariant] = useState<ProductVariant | null>(null);
  const [selectedImage, setSelectedImage] = useState(0);
  const [adding, setAdding] = useState(false);
  const [quantity, setQuantity] = useState(1);

  const { addItem } = useCartStore();
  const { isAuthenticated } = useAuthStore();

  useEffect(() => {
    catalogApi
      .get(`/products/${id}`)
      .then((res) => {
        // Catalog returns { data: product } or { product: ... }
        const p: Product = res.data?.data ?? res.data?.product ?? res.data;
        setProduct(p);
        if (p.variants?.length) setSelectedVariant(p.variants[0]);
      })
      .catch(() => setProduct(null))
      .finally(() => setLoading(false));
  }, [id]);

  const handleAddToCart = async () => {
    if (!isAuthenticated) {
      toast.error("Please sign in first");
      return;
    }
    if (!selectedVariant) {
      toast.error("Please select a variant");
      return;
    }
    setAdding(true);
    try {
      await orderApi.post("/cart/add", {
        product_variant_id: selectedVariant.id,
        quantity,
      });
      addItem({ product_variant_id: selectedVariant.id, quantity });
      toast.success(`${product?.title} added to cart!`);
    } catch {
      toast.error("Failed to add to cart");
    } finally {
      setAdding(false);
    }
  };

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-12">
          <Skeleton className="aspect-square rounded-2xl" />
          <div className="space-y-4">
            <Skeleton className="h-6 w-1/3" />
            <Skeleton className="h-10 w-3/4" />
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-12 w-1/2" />
            <Skeleton className="h-14 w-full" />
          </div>
        </div>
      </div>
    );
  }

  if (!product) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-20 text-center">
        <div className="text-5xl mb-4">📦</div>
        <h2 className="font-heading font-bold text-2xl text-foreground mb-2">Product Not Found</h2>
        <p className="text-muted-foreground mb-8">This product doesn&apos;t exist or has been removed.</p>
        <Link href="/products">
          <Button variant="outline" className="border-border/50">← Back to Products</Button>
        </Link>
      </div>
    );
  }

  const inStock = selectedVariant ? selectedVariant.inventory > 0 : false;
  // Normalize images to URL strings (catalog returns {url, altText} objects)
  const images: string[] = (product.images ?? []).map((img) =>
    typeof img === "string" ? img : (img as { url: string }).url
  );

  return (
    <div className="min-h-screen">
      {/* Breadcrumb */}
      <div className="border-b border-border/30 bg-card/30 backdrop-blur-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Link href="/" className="hover:text-foreground transition-colors">Home</Link>
            <span>/</span>
            <Link href="/products" className="hover:text-foreground transition-colors">Products</Link>
            <span>/</span>
            <span className="text-foreground truncate max-w-xs">{product.title}</span>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
        <Link
          href="/products"
          className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors mb-8"
        >
          <ChevronLeft className="w-4 h-4" />
          Back to Products
        </Link>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-12">
          {/* Main image — placeholder until MinIO/CDN configured */}
          <div className="relative aspect-square glass rounded-2xl overflow-hidden border border-border/50">
            <div className="w-full h-full flex flex-col items-center justify-center gap-3"
              style={{ background: "linear-gradient(135deg, oklch(0.2 0.04 285) 0%, oklch(0.15 0.03 285) 100%)" }}>
              <Package className="w-20 h-20 text-muted-foreground/20" />
              <span className="text-xs text-muted-foreground/50">Image preview unavailable</span>
            </div>
            {!inStock && (
              <div className="absolute inset-0 bg-black/50 flex items-center justify-center">
                <Badge variant="destructive" className="text-base px-4 py-2">Out of Stock</Badge>
              </div>
            )}
          </div>

          {/* Product Info */}
          <div className="space-y-6">
            {/* Brand & Title */}
            <div>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-primary uppercase tracking-wide">{product.brand}</span>
                <div className="flex gap-2">
                  <button className="p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-all">
                    <Heart className="w-4 h-4" />
                  </button>
                  <button className="p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-all">
                    <Share2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
              <h1 className="font-heading font-bold text-2xl md:text-3xl text-foreground leading-tight">
                {product.title}
              </h1>
            </div>

            {/* Rating */}
            <div className="flex items-center gap-2">
              <div className="flex">
                {[1,2,3,4,5].map(i => (
                  <Star key={i} className={`w-4 h-4 ${i <= 4 ? "fill-yellow-500 text-yellow-500" : "text-muted-foreground"}`} />
                ))}
              </div>
              <span className="text-sm text-muted-foreground">4.0 (128 reviews)</span>
            </div>

            {/* Price */}
            <div className="glass rounded-2xl p-5 border border-border/50">
              <div className="flex items-baseline gap-3">
                <span className="font-heading font-bold text-3xl text-foreground">
                  ₹{selectedVariant?.price.toLocaleString("en-IN") ?? "—"}
                </span>
              </div>
              <div className="flex items-center gap-2 mt-2">
                {inStock ? (
                  <><CheckCircle2 className="w-4 h-4 text-green-500" /><span className="text-sm text-green-500 font-medium">In Stock ({selectedVariant?.inventory} available)</span></>
                ) : (
                  <><XCircle className="w-4 h-4 text-destructive" /><span className="text-sm text-destructive font-medium">Out of Stock</span></>
                )}
              </div>
            </div>

            {/* Description */}
            <p className="text-muted-foreground leading-relaxed">{product.description}</p>

            {/* Variants */}
            {product.variants && product.variants.length > 1 && (
              <div>
                <h3 className="font-heading font-semibold text-foreground mb-3">Select Variant</h3>
                <div className="flex flex-wrap gap-2">
                  {product.variants.map((v) => (
                    <button
                      key={v.id}
                      onClick={() => setSelectedVariant(v)}
                      className={`px-4 py-2 rounded-xl text-sm border transition-all ${
                        selectedVariant?.id === v.id
                          ? "bg-primary/15 text-primary border-primary/40 glow-primary"
                          : v.inventory === 0
                          ? "border-border/30 text-muted-foreground/40 cursor-not-allowed line-through"
                          : "border-border/50 text-muted-foreground hover:text-foreground hover:border-border"
                      }`}
                      disabled={v.inventory === 0}
                    >
                      {v.sku} — ₹{v.price.toLocaleString()}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Quantity */}
            <div>
              <h3 className="font-heading font-semibold text-foreground mb-3">Quantity</h3>
              <div className="flex items-center gap-3">
                <button
                  onClick={() => setQuantity(Math.max(1, quantity - 1))}
                  className="w-10 h-10 rounded-xl border border-border/50 text-foreground hover:bg-secondary transition-all flex items-center justify-center font-bold"
                >
                  −
                </button>
                <span className="w-12 text-center font-semibold text-foreground text-lg">{quantity}</span>
                <button
                  onClick={() => setQuantity(Math.min(selectedVariant?.inventory ?? 10, quantity + 1))}
                  className="w-10 h-10 rounded-xl border border-border/50 text-foreground hover:bg-secondary transition-all flex items-center justify-center font-bold"
                >
                  +
                </button>
              </div>
            </div>

            {/* Specs */}
            {selectedVariant?.specifications && Object.keys(selectedVariant.specifications).length > 0 && (
              <div className="glass rounded-2xl p-5 border border-border/50">
                <h3 className="font-heading font-semibold text-foreground mb-3">Specifications</h3>
                <dl className="space-y-2">
                  {Object.entries(selectedVariant.specifications).map(([k, v]) => (
                    <div key={k} className="flex justify-between text-sm">
                      <dt className="text-muted-foreground">{k}</dt>
                      <dd className="text-foreground font-medium">{String(v)}</dd>
                    </div>
                  ))}
                </dl>
              </div>
            )}

            {/* Add to cart */}
            <Button
              id="add-to-cart-btn"
              onClick={handleAddToCart}
              disabled={!inStock || adding}
              size="lg"
              className="w-full h-14 text-base bg-primary hover:bg-primary/90 glow-primary gap-3 disabled:opacity-50"
            >
              {adding ? (
                <><Loader2 className="w-5 h-5 animate-spin" />Adding...</>
              ) : (
                <><ShoppingCart className="w-5 h-5" />{inStock ? "Add to Cart" : "Out of Stock"}</>
              )}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
