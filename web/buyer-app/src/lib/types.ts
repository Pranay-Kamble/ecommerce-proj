// Product types matching catalog service domain JSON tags
export interface ProductImage {
  url: string;
  altText: string;
  isPrimary: boolean;
}

export interface ProductVariant {
  id: string;            // json:"id" (PublicID)
  product_id?: string;
  sku: string;
  price: number;
  inventory: number;
  specifications?: Record<string, unknown>;
  images?: ProductImage[];
  createdAt?: string;
}

export interface Product {
  id: string;            // json:"id" (PublicID) — catalog uses "id" not "public_id"
  public_id?: string;    // alias for compatibility with search results
  title: string;
  slug: string;
  description: string;
  brand: string;
  highlights?: string[];
  dimensions?: Record<string, unknown>;
  category?: Category;
  seller?: Seller;
  images: ProductImage[] | string[]; // catalog returns objects; search returns strings
  variants: ProductVariant[];
  min_price?: number;
  max_price?: number;
  in_stock?: boolean;
  createdAt?: string;
  updatedAt?: string;
  // legacy snake_case aliases
  created_at?: string;
  updated_at?: string;
}

export interface Seller {
  id: string;
  name: string;
  description?: string;
  logoUrl?: string;
  supportEmail?: string;
}

export interface Category {
  id: string;            // json:"id" (PublicID)
  public_id?: string;    // alias
  name: string;
  slug?: string;
  path?: string;
  parentId?: string;
  children?: Category[];
}

export interface SearchResult {
  hits: SearchProduct[];
  total: number;
}

export interface SearchProduct {
  public_id: string;
  title: string;
  brand: string;
  description: string;
  category_name: string;
  seller_name: string;
  slug: string;
  images: string[];
  min_price: number;
  max_price: number;
  in_stock: boolean;
}

export interface CartItem {
  product_variant_id: string;
  quantity: number;
  price?: number;
  title?: string;
}

export interface Cart {
  user_id: string;
  items: CartItem[];
}

export interface Order {
  id: string;           // maps to PublicID (json:"id")
  user_id: string;
  total_amount: number;
  status: string;
  shipping_name: string;
  shipping_phone: string;
  shipping_address: string;
  shipping_city: string;
  shipping_state: string;
  shipping_zip: string;
  items: OrderItem[];
  created_at: string;
  updated_at: string;
}

export interface OrderItem {
  product_id: string;
  quantity: number;
  price: number;
}

export interface Address {
  title?: string;
  address_line: string;
  city: string;
  state: string;
  zip_code: string;
  is_default?: boolean;
}

export interface CustomerProfile {
  name: string;
  phone: string;
  addresses?: Address[];
}
