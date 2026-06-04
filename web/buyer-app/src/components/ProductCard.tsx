"use client";

import Link from "next/link";
import Image from "next/image";
import { ShoppingCart, Star, Package } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Product, SearchProduct } from "@/lib/types";
import { useCartStore, useAuthStore } from "@/lib/store";
import { orderApi } from "@/lib/api";
import toast from "react-hot-toast";

type ProductCardProps =
  | { product: Product; searchProduct?: never }
  | { product?: never; searchProduct: SearchProduct };

export function ProductCard({ product, searchProduct }: ProductCardProps) {
  const { addItem } = useCartStore();
  const { isAuthenticated } = useAuthStore();

  const id = product?.public_id ?? searchProduct!.public_id;
  const title = product?.title ?? searchProduct!.title;
  const brand = product?.brand ?? searchProduct!.brand;
  const images = product?.images ?? searchProduct?.images ?? [];
  const inStock = product
    ? product.variants?.some((v) => v.inventory > 0) ?? true
    : searchProduct!.in_stock;
  const minPrice = product
    ? Math.min(...(product.variants?.map((v) => v.price) ?? [0]))
    : searchProduct!.min_price;
  const maxPrice = product
    ? Math.max(...(product.variants?.map((v) => v.price) ?? [0]))
    : searchProduct!.max_price;

  const priceDisplay =
    minPrice === maxPrice
      ? `₹${minPrice.toLocaleString()}`
      : `₹${minPrice.toLocaleString()} – ₹${maxPrice.toLocaleString()}`;

  const handleAddToCart = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();

    if (!isAuthenticated) {
      toast.error("Please sign in to add items to cart");
      return;
    }

    const firstVariant = product?.variants?.[0];
    if (!firstVariant) {
      toast.error("No variant available");
      return;
    }

    try {
      await orderApi.post("/cart/add", {
        product_variant_id: firstVariant.public_id,
        quantity: 1,
      });
      addItem({ product_variant_id: firstVariant.public_id, quantity: 1 });
      toast.success("Added to cart!");
    } catch {
      toast.error("Failed to add to cart");
    }
  };

  return (
    <Link href={`/products/${id}`} className="group block">
      <div className="glass rounded-2xl overflow-hidden card-hover border border-border/50 group-hover:border-primary/30 transition-all duration-300">
        {/* Image */}
        <div className="relative aspect-[4/3] bg-secondary overflow-hidden">
          {images.length > 0 ? (
            <Image
              src={images[0]}
              alt={title}
              fill
              className="object-cover transition-transform duration-500 group-hover:scale-105"
              sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 33vw"
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center bg-secondary">
              <Package className="w-12 h-12 text-muted-foreground/30" />
            </div>
          )}
          {/* Badges */}
          <div className="absolute top-3 left-3 flex gap-2">
            {!inStock && (
              <Badge variant="destructive" className="text-xs font-medium">
                Out of Stock
              </Badge>
            )}
            {inStock && minPrice < 500 && (
              <Badge className="text-xs font-medium bg-primary/90">Great Value</Badge>
            )}
          </div>
          {/* Quick add overlay */}
          {inStock && (
            <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex items-end p-3">
              <Button
                onClick={handleAddToCart}
                size="sm"
                className="w-full bg-primary hover:bg-primary/90 text-primary-foreground gap-2"
              >
                <ShoppingCart className="w-4 h-4" />
                Quick Add
              </Button>
            </div>
          )}
        </div>

        {/* Info */}
        <div className="p-4">
          <p className="text-xs text-primary font-medium uppercase tracking-wide mb-1">{brand}</p>
          <h3 className="font-heading font-semibold text-foreground text-sm leading-tight line-clamp-2 mb-2 group-hover:text-primary transition-colors">
            {title}
          </h3>
          <div className="flex items-center gap-1 mb-3">
            {[1, 2, 3, 4, 5].map((i) => (
              <Star key={i} className={`w-3 h-3 ${i <= 4 ? "fill-yellow-500 text-yellow-500" : "text-muted-foreground"}`} />
            ))}
            <span className="text-xs text-muted-foreground ml-1">(4.0)</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="font-heading font-bold text-foreground">{priceDisplay}</span>
            {!inStock && (
              <span className="text-xs text-muted-foreground">Unavailable</span>
            )}
          </div>
        </div>
      </div>
    </Link>
  );
}
