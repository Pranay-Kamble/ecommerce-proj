import type { Metadata } from "next";
import { Inter, Outfit } from "next/font/google";
import "./globals.css";
import { Navbar } from "@/components/Navbar";
import { Footer } from "@/components/Footer";
import { Toaster } from "react-hot-toast";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
  weight: ["300", "400", "500", "600", "700", "800"],
});

const outfit = Outfit({
  subsets: ["latin"],
  variable: "--font-outfit",
  weight: ["400", "500", "600", "700", "800"],
});

export const metadata: Metadata = {
  title: "Nexus Store — Premium E-Commerce",
  description:
    "Discover premium products at Nexus Store. Shop the latest trends with fast delivery, easy returns, and a seamless checkout experience.",
  keywords: ["ecommerce", "shopping", "nexus store", "online shop"],
  openGraph: {
    title: "Nexus Store — Premium E-Commerce",
    description: "Discover premium products at Nexus Store.",
    type: "website",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className={`${inter.variable} ${outfit.variable} antialiased min-h-screen flex flex-col`}>
        <Toaster
          position="top-right"
          toastOptions={{
            style: {
              background: "oklch(0.13 0.008 265)",
              color: "oklch(0.96 0.005 265)",
              border: "1px solid oklch(0.22 0.012 265)",
              borderRadius: "0.75rem",
            },
          }}
        />
        <Navbar />
        <main className="flex-1">{children}</main>
        <Footer />
      </body>
    </html>
  );
}
