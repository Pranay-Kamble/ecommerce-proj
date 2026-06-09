"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ChevronRight, LayoutGrid, Loader2 } from "lucide-react";
import { Category } from "@/lib/types";
import { catalogApi } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

interface CategorySidebarProps {
  selectedCategory?: string;
  onSelectCategory: (id: string | undefined) => void;
}

export function CategorySidebar({ selectedCategory, onSelectCategory }: CategorySidebarProps) {
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    catalogApi
      .get("/categories")
      .then((r) => {
        const arr = r.data?.data ?? r.data?.categories ?? r.data;
        setCategories(Array.isArray(arr) ? arr : []);
      })
      .catch(() => setCategories([]))
      .finally(() => setLoading(false));
  }, []);

  return (
    <aside className="w-full">
      <div className="glass rounded-2xl p-5 border border-border/50">
        <h2 className="font-heading font-semibold text-foreground mb-4 flex items-center gap-2">
          <LayoutGrid className="w-4 h-4 text-primary" />
          Categories
        </h2>

        <nav className="space-y-1">
          <button
            onClick={() => onSelectCategory(undefined)}
            className={`w-full flex items-center justify-between px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-200 ${
              !selectedCategory
                ? "bg-primary/15 text-primary"
                : "text-muted-foreground hover:text-foreground hover:bg-secondary"
            }`}
          >
            <span>All Products</span>
            {!selectedCategory && <ChevronRight className="w-4 h-4" />}
          </button>

          {loading
            ? Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full rounded-xl" />
              ))
            : categories.map((cat) => (
                <button
                  key={cat.id}
                  onClick={() => onSelectCategory(cat.id)}
                  className={`w-full flex items-center justify-between px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-200 ${
                    selectedCategory === cat.id
                      ? "bg-primary/15 text-primary"
                      : "text-muted-foreground hover:text-foreground hover:bg-secondary"
                  }`}
                >
                  <span className="truncate">{cat.name}</span>
                  {selectedCategory === cat.public_id && <ChevronRight className="w-4 h-4 shrink-0" />}
                </button>
              ))}
        </nav>
      </div>
    </aside>
  );
}
