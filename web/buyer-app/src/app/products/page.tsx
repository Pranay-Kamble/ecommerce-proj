import { Suspense } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import ProductsContent from "./ProductsContent";

export const metadata = {
  title: "Products — Nexus Store",
  description: "Browse all products. Search by name, filter by category, price, and availability.",
};

function LoadingFallback() {
  return (
    <div className="min-h-screen">
      <div className="border-b border-border/30 bg-card/30 backdrop-blur-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <Skeleton className="h-10 w-64 mb-6" />
          <Skeleton className="h-11 w-full" />
        </div>
      </div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex gap-8">
          <div className="hidden md:block w-56 shrink-0">
            <Skeleton className="h-64 w-full rounded-2xl" />
          </div>
          <div className="flex-1 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="aspect-[4/3] rounded-2xl" />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

export default function ProductsPage() {
  return (
    <Suspense fallback={<LoadingFallback />}>
      <ProductsContent />
    </Suspense>
  );
}
