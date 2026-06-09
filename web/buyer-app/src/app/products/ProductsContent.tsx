"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { Search, SlidersHorizontal, X, Loader2 } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { ProductCard } from "@/components/ProductCard";
import { CategorySidebar } from "@/components/CategorySidebar";
import { searchApi, catalogApi } from "@/lib/api";
import { Product, SearchProduct } from "@/lib/types";
import { Button } from "@/components/ui/button";

export default function ProductsContent() {
  const searchParams = useSearchParams();
  const router = useRouter();

  const [query, setQuery] = useState(searchParams.get("q") ?? "");
  const [selectedCategory, setSelectedCategory] = useState<string | undefined>(
    searchParams.get("category") ?? undefined
  );
  const [minPrice, setMinPrice] = useState(searchParams.get("min_price") ?? "");
  const [maxPrice, setMaxPrice] = useState(searchParams.get("max_price") ?? "");
  const [inStockOnly, setInStockOnly] = useState(false);
  const [results, setResults] = useState<(Product | SearchProduct)[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [showFilters, setShowFilters] = useState(false);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const searchRef = useRef<HTMLDivElement>(null);

  // Close suggestions on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (searchRef.current && !searchRef.current.contains(e.target as Node)) {
        setShowSuggestions(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  const fetchProducts = useCallback(async () => {
    setLoading(true);
    try {
      if (query.trim()) {
        const params: Record<string, string> = { q: query };
        if (selectedCategory) params.category = selectedCategory;
        if (minPrice) params.min_price = minPrice;
        if (maxPrice) params.max_price = maxPrice;
        if (inStockOnly) params.in_stock = "true";

        const res = await searchApi.get("/products", { params });
        const data = res.data;
        setResults(data.hits ?? data.products ?? []);
        setTotal(data.total ?? data.hits?.length ?? 0);
      } else {
        const params: Record<string, string> = {};
        if (selectedCategory) params.category_id = selectedCategory;

        const res = await catalogApi.get("/products", { params });
        // Catalog API returns { data: [...] }
        const raw = res.data?.data ?? res.data?.products ?? res.data;
        const products = Array.isArray(raw) ? raw : [];
        setResults(products);
        setTotal(products.length);
      }
    } catch {
      setResults([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, [query, selectedCategory, minPrice, maxPrice, inStockOnly]);

  // Autocomplete suggestions
  const fetchSuggestions = useCallback(async (q: string) => {
    if (q.trim().length < 2) { setSuggestions([]); return; }
    try {
      const res = await searchApi.get("/products", { params: { q, size: "5" } });
      const hits: SearchProduct[] = res.data?.hits ?? [];
      setSuggestions(hits.map((h) => h.title).filter(Boolean).slice(0, 5));
    } catch {
      setSuggestions([]);
    }
  }, []);

  useEffect(() => {
    const timer = setTimeout(fetchProducts, 400);
    return () => clearTimeout(timer);
  }, [fetchProducts]);

  // Separate debounce for autocomplete (faster, 200ms)
  useEffect(() => {
    const timer = setTimeout(() => fetchSuggestions(query), 200);
    return () => clearTimeout(timer);
  }, [query, fetchSuggestions]);

  const handleCategoryChange = (id: string | undefined) => {
    setSelectedCategory(id);
    const params = new URLSearchParams(searchParams.toString());
    if (id) params.set("category", id);
    else params.delete("category");
    router.push(`/products?${params.toString()}`);
  };

  const clearFilters = () => {
    setQuery("");
    setSelectedCategory(undefined);
    setMinPrice("");
    setMaxPrice("");
    setInStockOnly(false);
    router.push("/products");
  };

  const hasActiveFilters = query || selectedCategory || minPrice || maxPrice || inStockOnly;

  const isSearchProduct = (p: Product | SearchProduct): p is SearchProduct =>
    !("variants" in p);

  return (
    <div className="min-h-screen">
      {/* Header */}
      <div className="border-b border-border/30 bg-card/30 backdrop-blur-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <h1 className="font-heading font-bold text-3xl md:text-4xl text-foreground mb-6">
            {selectedCategory ? "Category Products" : query ? `Results for "${query}"` : "All Products"}
          </h1>

          <div className="flex gap-3">
            <div ref={searchRef} className="relative flex-1">
              <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground z-10" />
              <Input
                id="search-input"
                placeholder="Search products, brands, categories..."
                value={query}
                onChange={(e) => { setQuery(e.target.value); setShowSuggestions(true); }}
                onFocus={() => suggestions.length > 0 && setShowSuggestions(true)}
                className="pl-10 h-11 bg-secondary border-border/50 focus:border-primary/50"
                autoComplete="off"
              />
              {query && (
                <button
                  onClick={() => { setQuery(""); setSuggestions([]); setShowSuggestions(false); }}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground z-10"
                >
                  <X className="w-4 h-4" />
                </button>
              )}
              {/* Autocomplete dropdown */}
              {showSuggestions && suggestions.length > 0 && (
                <div className="absolute top-full left-0 right-0 mt-1 glass-strong rounded-xl border border-border/60 shadow-xl z-50 overflow-hidden">
                  {suggestions.map((s, i) => (
                    <button
                      key={i}
                      onMouseDown={(e) => e.preventDefault()}
                      onClick={() => { setQuery(s); setShowSuggestions(false); }}
                      className="w-full px-4 py-2.5 text-left text-sm text-foreground hover:bg-secondary flex items-center gap-3 transition-colors border-b border-border/20 last:border-0"
                    >
                      <Search className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
                      {s}
                    </button>
                  ))}
                </div>
              )}
            </div>
            <Button
              variant="outline"
              onClick={() => setShowFilters(!showFilters)}
              className="gap-2 border-border/50 md:hidden"
            >
              <SlidersHorizontal className="w-4 h-4" />
              Filters
            </Button>
          </div>

          {showFilters && (
            <div className="flex flex-wrap gap-3 mt-4 md:hidden">
              <Input
                placeholder="Min price"
                value={minPrice}
                onChange={(e) => setMinPrice(e.target.value)}
                type="number"
                className="w-28 h-9 text-sm bg-secondary border-border/50"
              />
              <Input
                placeholder="Max price"
                value={maxPrice}
                onChange={(e) => setMaxPrice(e.target.value)}
                type="number"
                className="w-28 h-9 text-sm bg-secondary border-border/50"
              />
              <button
                onClick={() => setInStockOnly(!inStockOnly)}
                className={`px-3 h-9 rounded-lg text-sm border transition-all ${
                  inStockOnly
                    ? "bg-primary/15 text-primary border-primary/30"
                    : "border-border/50 text-muted-foreground hover:text-foreground"
                }`}
              >
                In Stock Only
              </button>
            </div>
          )}

          {hasActiveFilters && (
            <div className="flex items-center gap-2 mt-3 flex-wrap">
              <span className="text-xs text-muted-foreground">Active filters:</span>
              {query && <Badge variant="secondary" className="text-xs">Search: {query}</Badge>}
              {selectedCategory && <Badge variant="secondary" className="text-xs">Category filter</Badge>}
              {minPrice && <Badge variant="secondary" className="text-xs">Min: ₹{minPrice}</Badge>}
              {maxPrice && <Badge variant="secondary" className="text-xs">Max: ₹{maxPrice}</Badge>}
              {inStockOnly && <Badge variant="secondary" className="text-xs">In Stock</Badge>}
              <button onClick={clearFilters} className="text-xs text-primary hover:underline">
                Clear all
              </button>
            </div>
          )}
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex gap-8">
          {/* Desktop sidebar */}
          <div className="hidden md:block w-56 shrink-0">
            <CategorySidebar selectedCategory={selectedCategory} onSelectCategory={handleCategoryChange} />

            <div className="glass rounded-2xl p-5 border border-border/50 mt-4">
              <h3 className="font-heading font-semibold text-foreground mb-4 text-sm">Price Range</h3>
              <div className="space-y-3">
                <Input
                  placeholder="Min price (₹)"
                  value={minPrice}
                  onChange={(e) => setMinPrice(e.target.value)}
                  type="number"
                  className="h-9 text-sm bg-secondary border-border/50"
                />
                <Input
                  placeholder="Max price (₹)"
                  value={maxPrice}
                  onChange={(e) => setMaxPrice(e.target.value)}
                  type="number"
                  className="h-9 text-sm bg-secondary border-border/50"
                />
                <button
                  onClick={() => setInStockOnly(!inStockOnly)}
                  className={`w-full px-3 h-9 rounded-xl text-sm border transition-all ${
                    inStockOnly
                      ? "bg-primary/15 text-primary border-primary/30"
                      : "border-border/50 text-muted-foreground hover:text-foreground hover:border-border"
                  }`}
                >
                  {inStockOnly ? "✓ " : ""}In Stock Only
                </button>
              </div>
            </div>
          </div>

          {/* Product grid */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center justify-between mb-6">
              <p className="text-sm text-muted-foreground">
                {loading ? (
                  <span className="flex items-center gap-2">
                    <Loader2 className="w-3 h-3 animate-spin" /> Loading...
                  </span>
                ) : `${total} product${total !== 1 ? "s" : ""} found`}
              </p>
            </div>

            {loading ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                {Array.from({ length: 6 }).map((_, i) => (
                  <div key={i} className="glass rounded-2xl overflow-hidden border border-border/50">
                    <Skeleton className="aspect-[4/3] w-full" />
                    <div className="p-4 space-y-2">
                      <Skeleton className="h-3 w-1/3" />
                      <Skeleton className="h-5 w-full" />
                      <Skeleton className="h-4 w-1/2" />
                    </div>
                  </div>
                ))}
              </div>
            ) : results.length === 0 ? (
              <div className="text-center py-24 glass rounded-2xl border border-border/50">
                <div className="text-5xl mb-4">🔍</div>
                <h3 className="font-heading font-semibold text-foreground text-xl mb-2">No products found</h3>
                <p className="text-muted-foreground text-sm mb-6">
                  Try adjusting your search or filters.
                </p>
                <Button onClick={clearFilters} variant="outline" className="border-border/50">
                  Clear Filters
                </Button>
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                {results.map((p) =>
                  isSearchProduct(p) ? (
                    <ProductCard key={p.public_id} searchProduct={p} />
                  ) : (
                    <ProductCard key={(p as Product).id} product={p as Product} />
                  )
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
