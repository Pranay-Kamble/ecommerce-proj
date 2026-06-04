import Link from "next/link";
import { Zap, ExternalLink } from "lucide-react";

export function Footer() {
  return (
    <footer className="border-t border-border/50 bg-card/50 backdrop-blur-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          {/* Brand */}
          <div className="md:col-span-1">
            <Link href="/" className="flex items-center gap-2 mb-4">
              <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
                <Zap className="w-4 h-4 text-primary-foreground" />
              </div>
              <span className="font-heading font-bold text-xl gradient-text">Nexus</span>
            </Link>
            <p className="text-sm text-muted-foreground leading-relaxed">
              Premium products, seamless checkout, and reliable delivery — all in one place.
            </p>
            <div className="flex items-center gap-3 mt-4">
              <a href="https://github.com" target="_blank" rel="noreferrer" className="p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-all">
                <ExternalLink className="w-4 h-4" />
              </a>
              <a href="https://twitter.com" target="_blank" rel="noreferrer" className="p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-all">
                <ExternalLink className="w-4 h-4" />
              </a>
              <a href="https://instagram.com" target="_blank" rel="noreferrer" className="p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-all">
                <ExternalLink className="w-4 h-4" />
              </a>
            </div>
          </div>

          {/* Shop */}
          <div>
            <h3 className="font-heading font-semibold text-foreground mb-4">Shop</h3>
            <ul className="space-y-2">
              {["All Products", "New Arrivals", "Best Sellers", "Sale"].map((item) => (
                <li key={item}>
                  <Link href="/products" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
                    {item}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* Account */}
          <div>
            <h3 className="font-heading font-semibold text-foreground mb-4">Account</h3>
            <ul className="space-y-2">
              {[
                { label: "Sign In", href: "/auth/login" },
                { label: "Register", href: "/auth/register" },
                { label: "My Orders", href: "/orders" },
                { label: "Cart", href: "/cart" },
              ].map((item) => (
                <li key={item.label}>
                  <Link href={item.href} className="text-sm text-muted-foreground hover:text-foreground transition-colors">
                    {item.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* Legal */}
          <div>
            <h3 className="font-heading font-semibold text-foreground mb-4">Legal</h3>
            <ul className="space-y-2">
              {["Privacy Policy", "Terms of Service", "Cookie Policy", "Contact Us"].map((item) => (
                <li key={item}>
                  <a href="#" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
                    {item}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        </div>

        <div className="border-t border-border/50 mt-10 pt-6 flex flex-col md:flex-row items-center justify-between gap-4">
          <p className="text-xs text-muted-foreground">
            © {new Date().getFullYear()} Nexus Store. All rights reserved.
          </p>
          <p className="text-xs text-muted-foreground">
            Built with Next.js, Go microservices, and ❤️
          </p>
        </div>
      </div>
    </footer>
  );
}
