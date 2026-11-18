// Bentuk data sesuai backend. Semua uang berupa string desimal, id berupa UUID.
export type ApiMeta = { page: number; limit: number; total: number };
export type Paginated<T> = { data: T[]; meta: ApiMeta };

export type Role = "customer" | "owner" | "staff";

export type ApiUser = {
  id: string;
  name: string;
  email: string;
  phone: string | null;
  role: Role;
  status: "active" | "inactive";
  email_verified_at: string | null;
  created_at: string;
};

export type AuthResult = {
  access_token: string;
  token_type: string;
  expires_in: number;
  user: ApiUser;
};

export type ApiProduct = {
  id: string;
  category_id: string | null;
  name: string;
  description: string | null;
  price: string;
  discount_percent: number;
  weight_gram: number;
  availability_type: "ready" | "preorder" | string;
  po_estimate_days: number | null;
  status: "active" | "inactive" | string;
  primary_image_url?: string | null;
  /** Hanya di daftar produk. */
  total_stock?: number;
  variant_count?: number;
};

export type ApiVariant = {
  id: string;
  product_id: string;
  sku: string;
  size: string;
  color: string | null;
  stock: number;
};

export type ApiImage = {
  id: string;
  product_id: string;
  image_url: string;
  sort_order: number;
  is_primary: boolean;
};

export type ApiProductDetail = ApiProduct & {
  variants: ApiVariant[];
  images: ApiImage[];
};

export type ApiCategory = { id: string; name: string; is_active: boolean };

export type ApiCartItem = {
  id: string;
  variant_id: string;
  product_id: string;
  product_name: string;
  sku: string;
  size: string;
  color: string | null;
  unit_price: string;
  quantity: number;
  stock: number;
  subtotal: string;
  primary_image_url: string | null;
};
export type ApiCart = { items: ApiCartItem[]; subtotal: string };

export type OrderStatus = "pending_payment" | "paid" | "processing" | "shipped" | "completed" | "cancelled";

export type ApiOrder = {
  id: string;
  order_number: string;
  user_id: string;
  recipient_name: string;
  recipient_phone: string;
  shipping_address: string;
  shipping_city: string | null;
  shipping_kelurahan: string;
  shipping_district: string | null;
  shipping_subdistrict_id: number | null;
  shipping_zip_code: string | null;
  subtotal: string;
  shipping_cost: string;
  total_amount: string;
  order_status: OrderStatus;
  courier_code: string | null;
  courier_service: string | null;
  shipping_etd: string | null;
  shipping_weight_gram: number | null;
  created_at: string;
  updated_at: string;
  payment_due_at: string | null;
  cancel_reason: string | null;
};

/** Order di daftar (dengan ringkasan item & pembayaran). */
export type ApiOrderListItem = ApiOrder & {
  item_count: number;
  payment_status: PaymentStatus | null;
  has_proof: boolean;
  customer_email: string;
};

export type ApiOrderItem = {
  id: string;
  order_id: string;
  product_id: string;
  variant_id: string;
  product_name: string;
  size: string;
  unit_price: string;
  quantity: number;
  subtotal: string;
  image_url: string | null;
};

export type PaymentStatus = "menunggu" | "valid" | "ditolak";

export type ApiPayment = {
  id: string;
  order_id: string;
  method: "qris" | "bank_transfer" | "ewallet" | string;
  amount: string;
  unique_code: number;
  proof_type: string | null;
  has_proof: boolean;
  status: PaymentStatus;
  paid_at: string | null;
  reviewed_by: string | null;
  reviewed_at: string | null;
  created_at: string;
};

export type ApiAdminPayment = ApiPayment & {
  order_number: string;
  order_status: OrderStatus;
  order_total: string;
  payment_due_at: string | null;
  customer_name: string;
  customer_email: string;
};

export type ApiShipment = {
  id: string;
  order_id: string;
  /** Hanya dibaca: kurir pesanan saat checkout (kode + layanan). */
  courier_name: string;
  tracking_number: string | null;
  coverage_status: string;
  shipping_status: "pending" | "shipped" | "delivered" | "returned" | string;
  updated_at: string;
};

export type ApiOrderDetail = ApiOrder & {
  items: ApiOrderItem[];
  payment: ApiPayment | null;
  shipment: ApiShipment | null;
};

/** Item wilayah dari RajaOngkir (provinsi/kota/kecamatan/kelurahan). */
export type ApiRegion = { id: number; name: string; zip_code?: string; [key: string]: unknown };

export type ApiShippingQuote = {
  courier_code: string;
  courier_name: string;
  service: string;
  description: string;
  cost: string;
  etd: string;
  weight_gram: number;
};

