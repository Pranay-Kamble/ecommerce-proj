import Link from "next/link";
import { ArrowRight, Zap, Shield, Truck, Star, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { catalogApi } from "@/lib/api";
import { Product, Category } from "@/lib/types";
import { ProductCard } from "@/components/ProductCard";

async function getFeaturedProducts(): Promise<Product[]> {
  try {
    const res = await catalogApi.get("/products?limit=8");
    // Catalog API returns { data: [...] }
    const arr = res.data?.data ?? res.data?.products ?? res.data;
    return Array.isArray(arr) ? arr : [];
  } catch {
    return [];
  }
}

async function getCategories(): Promise<Category[]> {
  try {
    const res = await catalogApi.get("/categories");
    // Catalog API returns { data: [...] }
    const arr = res.data?.data ?? res.data?.categories ?? res.data;
    return Array.isArray(arr) ? arr : [];
  } catch {
    return [];
  }
}

const features = [
  {
    icon: Zap,
    title: "Lightning Fast",
    description: "Blazing-fast checkout with one-click Stripe payments.",
  },
  {
    icon: Shield,
    title: "Secure Shopping",
    description: "RSA-JWT authentication & end-to-end encryption on all transactions.",
  },
  {
    icon: Truck,
    title: "Fast Delivery",
    description: "Nationwide delivery with real-time tracking on every order.",
  },
];

const categoryIcons = ["👗", "👟", "📱", "💻", "🏠", "⌚", "🎮", "📚"];

export default async function HomePage() {
  const [products, categories] = await Promise.all([getFeaturedProducts(), getCategories()]);

  return (
    <div className="min-h-screen">
      {/* Hero Section */}
      <section className="relative hero-gradient overflow-hidden">
        {/* Background grid */}
        <div
          className="absolute inset-0 opacity-5"
          style={{
            backgroundImage:
              "linear-gradient(oklch(0.65 0.22 285) 1px, transparent 1px), linear-gradient(90deg, oklch(0.65 0.22 285) 1px, transparent 1px)",
            backgroundSize: "60px 60px",
          }}
        />
        {/* Decorative blobs */}
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-primary/5 rounded-full blur-3xl" />
        <div className="absolute bottom-1/4 right-1/4 w-80 h-80 bg-violet-500/5 rounded-full blur-3xl" />

        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24 md:py-36 text-center">
          <Badge className="mb-6 bg-primary/15 text-primary border-primary/30 hover:bg-primary/20 cursor-default">
            <Zap className="w-3 h-3 mr-1" />
            Powered by Microservices
          </Badge>

          <h1 className="font-heading font-extrabold text-4xl sm:text-5xl md:text-7xl leading-tight mb-6">
            Shop the Future,{" "}
            <span className="gradient-text">Delivered Today</span>
          </h1>

          <p className="text-lg md:text-xl text-muted-foreground max-w-2xl mx-auto mb-10 leading-relaxed">
            Discover thousands of premium products with lightning-fast checkout,
            real-time inventory, and AI-powered search built on Go microservices.
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
            <Link href="/products">
              <Button size="lg" className="bg-primary hover:bg-primary/90 glow-primary gap-2 text-base px-8 h-12">
                Shop Now
                <ArrowRight className="w-5 h-5" />
              </Button>
            </Link>
            <Link href="/products">
              <Button
                size="lg"
                variant="outline"
                className="border-border/60 text-muted-foreground hover:text-foreground gap-2 text-base px-8 h-12"
              >
                <Search className="w-5 h-5" />
                Search Products
              </Button>
            </Link>
          </div>

          {/* Stats */}
          <div className="flex flex-wrap justify-center gap-8 mt-16">
            {[
              { value: "10K+", label: "Products" },
              { value: "50K+", label: "Customers" },
              { value: "99.9%", label: "Uptime" },
              { value: "4.9★", label: "Rating" },
            ].map((stat) => (
              <div key={stat.label} className="text-center">
                <div className="font-heading font-bold text-2xl text-foreground">{stat.value}</div>
                <div className="text-sm text-muted-foreground">{stat.label}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="py-20 border-y border-border/30">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {features.map((feature) => {
              const Icon = feature.icon;
              return (
                <div key={feature.title} className="glass rounded-2xl p-6 border border-border/50 flex gap-4 items-start">
                  <div className="w-10 h-10 rounded-xl bg-primary/15 flex items-center justify-center shrink-0">
                    <Icon className="w-5 h-5 text-primary" />
                  </div>
                  <div>
                    <h3 className="font-heading font-semibold text-foreground mb-1">{feature.title}</h3>
                    <p className="text-sm text-muted-foreground leading-relaxed">{feature.description}</p>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </section>

      {/* Categories */}
      {categories.length > 0 && (
        <section className="py-20">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex items-end justify-between mb-10">
              <div>
                <p className="text-primary font-medium text-sm uppercase tracking-wide mb-2">Browse</p>
                <h2 className="font-heading font-bold text-3xl md:text-4xl text-foreground">
                  Shop by Category
                </h2>
              </div>
              <Link href="/products" className="text-sm text-muted-foreground hover:text-primary transition-colors flex items-center gap-1">
                View all <ArrowRight className="w-4 h-4" />
              </Link>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
              {categories.slice(0, 8).map((cat, i) => (
                <Link
                  key={cat.public_id}
                  href={`/products?category=${cat.public_id}`}
                  className="group glass rounded-2xl p-5 border border-border/50 hover:border-primary/40 card-hover text-center transition-all duration-300"
                >
                  <div className="text-3xl mb-3">{categoryIcons[i] ?? "🛍️"}</div>
                  <h3 className="font-heading font-semibold text-foreground text-sm group-hover:text-primary transition-colors">
                    {cat.name}
                  </h3>
                </Link>
              ))}
            </div>
          </div>
        </section>
      )}

      {/* Featured Products */}
      <section className="py-20 border-t border-border/30">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-end justify-between mb-10">
            <div>
              <p className="text-primary font-medium text-sm uppercase tracking-wide mb-2">Curated</p>
              <h2 className="font-heading font-bold text-3xl md:text-4xl text-foreground">
                Featured Products
              </h2>
            </div>
            <Link href="/products" className="text-sm text-muted-foreground hover:text-primary transition-colors flex items-center gap-1">
              View all <ArrowRight className="w-4 h-4" />
            </Link>
          </div>

          {products.length === 0 ? (
            <div className="text-center py-20 glass rounded-2xl border border-border/50">
              <div className="text-5xl mb-4">🛍️</div>
              <h3 className="font-heading font-semibold text-foreground mb-2">No Products Yet</h3>
              <p className="text-muted-foreground text-sm">
                Add products through the Seller Dashboard to see them here.
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
              {products.map((product) => (
                <ProductCard key={product.id} product={product} />
              ))}
            </div>
          )}
        </div>
      </section>

      {/* CTA Banner */}
      <section className="py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="relative glass rounded-3xl p-10 md:p-16 border border-primary/20 overflow-hidden text-center">
            <div className="absolute inset-0 bg-gradient-to-r from-primary/5 via-transparent to-violet-500/5" />
            <div className="relative">
              <div className="flex justify-center mb-4">
                {[1,2,3,4,5].map(i => (
                  <Star key={i} className="w-5 h-5 fill-yellow-500 text-yellow-500" />
                ))}
              </div>
              <h2 className="font-heading font-bold text-3xl md:text-4xl text-foreground mb-4">
                Join 50,000+ Happy Shoppers
              </h2>
              <p className="text-muted-foreground max-w-xl mx-auto mb-8">
                Create your account today and get access to exclusive deals,
                order tracking, and PDF invoices for every purchase.
              </p>
              <Link href="/auth/register">
                <Button size="lg" className="bg-primary hover:bg-primary/90 glow-primary gap-2 px-10 h-12">
                  Get Started Free
                  <ArrowRight className="w-5 h-5" />
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
