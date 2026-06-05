// Product types matching backend domain models
export interface ProductVariant {
  id: string;
  public_id: string;
  product_id: string;
  sku: string;
  price: number;
  compare_price?: number;
  inventory: number;
  specifications?: Record<string, string>;
  created_at: string;
}

export interface Product {
  id: string;
  public_id: string;
  title: string;
  slug: string;
  description: string;
  brand: string;
  category_id: string;
  category_name?: string;
  seller_id: string;
  seller_name?: string;
  images: string[];
  variants: ProductVariant[];
  min_price?: number;
  max_price?: number;
  in_stock?: boolean;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: string;
  public_id: string;
  name: string;
  slug: string;
  parent_id?: string;
  path?: string;
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
