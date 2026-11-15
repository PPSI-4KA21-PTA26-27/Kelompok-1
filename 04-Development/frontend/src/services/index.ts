import { api, fetchBlob } from "@/lib/api";
import type {
  ApiAdminPayment,
  ApiAnalytics,
  ApiBanner,
  ApiCart,
  ApiCategory,
  ApiImage,
  ApiOrder,
  ApiOrderDetail,
  ApiOrderListItem,
  ApiPayment,
  ApiPost,
  ApiPostDetail,
  ApiProduct,
  ApiProductDetail,
  ApiRegion,
  ApiShipment,
  ApiShippingQuote,
  ApiSiteSettings,
  ApiSummary,
  ApiUser,
  ApiVariant,
  AuthResult,
  BannerInput,
  BannerPlacement,
  PostInput,
} from "@/types/api";

export type ProductQuery = {
  q?: string;
  category_id?: string;
  min_price?: number;
  max_price?: number;
  availability_type?: string;
  /** Hanya produk yang dibuat N hari terakhir. */
  new_days?: number;
  sort?: string;
  page?: number;
  limit?: number;
};

export const authService = {
  login: (email: string, password: string) =>
    api.post<AuthResult>("/auth/login", { email, password }, { auth: false }),
  register: (input: { name: string; email: string; password: string; phone?: string | null }) =>
    api.post<AuthResult>("/auth/register", input, { auth: false }),
  verifyEmail: (token: string) => api.post<ApiUser>("/auth/verify-email", { token }, { auth: false }),
  resendVerification: () => api.post<void>("/auth/verification/resend"),
  forgotPassword: (email: string) => api.post<void>("/auth/password/forgot", { email }, { auth: false }),
  resetPassword: (token: string, password: string) =>
    api.post<void>("/auth/password/reset", { token, password }, { auth: false }),
  me: () => api.get<ApiUser>("/users/me"),
  updateMe: (input: { name?: string; phone?: string | null }) => api.patch<ApiUser>("/users/me", input),
  changePassword: (input: { current_password: string; new_password: string }) =>
    api.put<void>("/users/me/password", input),
};

export const productService = {
  list: (query?: ProductQuery) => api.list<ApiProduct>("/products", query, { auth: false }),
  get: (id: string) => api.get<ApiProductDetail>("/products/" + id, undefined, { auth: false }),
};

export const categoryService = {
  list: () => api.list<ApiCategory>("/categories", { limit: 100 }, { auth: false }),
};

export const bannerService = {
  /** Banner yang sedang tayang untuk satu penempatan. */
  live: (placement: BannerPlacement) => api.get<ApiBanner[]>("/banners", { placement }, { auth: false }),
};

export const settingService = {
  get: () => api.get<ApiSiteSettings>("/site-settings", undefined, { auth: false }),
};

export const postService = {
  list: (query?: { tag?: string; page?: number; limit?: number }) =>
    api.list<ApiPost>("/posts", query, { auth: false }),
  get: (slug: string) => api.get<ApiPostDetail>("/posts/" + encodeURIComponent(slug), undefined, { auth: false }),
};

export const cartService = {
  get: () => api.get<ApiCart>("/cart"),
  add: (variant_id: string, quantity: number) => api.post<unknown>("/cart/items", { variant_id, quantity }),
  update: (id: string, quantity: number) => api.patch<unknown>("/cart/items/" + id, { quantity }),
  remove: (id: string) => api.del("/cart/items/" + id),
  clear: () => api.del("/cart"),
};

export const regionService = {
  provinces: () => api.get<ApiRegion[]>("/regions/provinces"),
  cities: (provinceId: number) => api.get<ApiRegion[]>("/regions/provinces/" + provinceId + "/cities"),
  districts: (cityId: number) => api.get<ApiRegion[]>("/regions/cities/" + cityId + "/districts"),
  subdistricts: (districtId: number) => api.get<ApiRegion[]>("/regions/districts/" + districtId + "/subdistricts"),
  shippingCost: (subdistrictId: number, cartItemIds?: string[]) =>
    api.get<ApiShippingQuote>("/shipping/costs", {
      subdistrict_id: subdistrictId,
      cart_item_ids: cartItemIds?.length ? cartItemIds.join(",") : undefined,
    }),
};

export type CheckoutInput = {
  recipient_name: string;
  recipient_phone: string;
  shipping_address: string;
  rt?: string;
  rw?: string;
  province_id: number;
  city_id: number;
  district_id: number;
  subdistrict_id: number;
  cart_item_ids: string[];
};

export const orderService = {
  listMine: (query?: { page?: number; limit?: number; order_status?: string }) =>
    api.list<ApiOrderListItem>("/orders", query),
  get: (id: string) => api.get<ApiOrderDetail>("/orders/" + id),
  create: (input: CheckoutInput) => api.post<ApiOrderDetail>("/orders", input),
  cancel: (id: string, reason?: string) => api.post<ApiOrder>("/orders/" + id + "/cancel", reason ? { reason } : undefined),
  complete: (id: string) => api.post<ApiOrder>("/orders/" + id + "/complete"),
};

export const paymentService = {
  /** Tagihan QRIS (order lama yang belum punya tagihan). */
  create: (orderId: string) => api.post<ApiPayment>("/orders/" + orderId + "/payment", { method: "qris" }),
  /** Bukti bayar dikirim multipart, field "file" (JPG/PNG/WebP/PDF, maks 2 MB). */
  uploadProof: (orderId: string, file: File) => {
    const fd = new FormData();
    fd.append("file", file);
    return api.put<ApiPayment>("/orders/" + orderId + "/payment/proof", fd);
  },
  proof: (orderId: string) => fetchBlob("/orders/" + orderId + "/payment/proof"),
};

// ===== Admin & staff =====

export type ProductInput = {
  category_id: string | null;
  name: string;
  description: string | null;
  price: string;
  discount_percent: number;
  weight_gram: number;
  availability_type: "ready" | "preorder";
  po_estimate_days: number | null;
  status: "active" | "inactive";
};

export type VariantInput = { sku: string; size: string; color: string | null; stock: number };

export const adminService = {
  summary: () => api.get<ApiSummary>("/admin/summary"),
  analytics: (days: 7 | 30 | 90) => api.get<ApiAnalytics>("/admin/analytics", { days }),

  orders: (query?: { page?: number; limit?: number; order_status?: string; q?: string }) =>
    api.list<ApiOrderListItem>("/admin/orders", query),
  updateOrderStatus: (
    id: string,
    input: { order_status: string; tracking_number?: string; reason?: string }
  ) => api.put<ApiOrder>("/admin/orders/" + id + "/status", input),
  upsertShipment: (
    orderId: string,
    input: { tracking_number: string | null; coverage_status: string; shipping_status: string }
  ) => api.put<ApiShipment>("/orders/" + orderId + "/shipment", input),

  payments: (query?: { status?: string; q?: string; page?: number; limit?: number }) =>
    api.list<ApiAdminPayment>("/admin/payments", query),
  reviewPayment: (paymentId: string, status: "valid" | "ditolak") =>
    api.post<ApiPayment>("/admin/payments/" + paymentId + "/review", { status }),

  products: (query?: ProductQuery & { status?: string }) => api.list<ApiProduct>("/products", query),
  product: (id: string) => api.get<ApiProductDetail>("/products/" + id),
  createProduct: (input: ProductInput) => api.post<ApiProduct>("/products", input),
  updateProduct: (id: string, input: Partial<ProductInput>) => api.patch<ApiProduct>("/products/" + id, input),
  deleteProduct: (id: string) => api.del("/products/" + id),
  createVariant: (productId: string, input: VariantInput) =>
    api.post<ApiVariant>("/products/" + productId + "/variants", input),
  updateVariant: (productId: string, variantId: string, input: Partial<VariantInput>) =>
    api.patch<ApiVariant>("/products/" + productId + "/variants/" + variantId, input),
  deleteVariant: (productId: string, variantId: string) => api.del("/products/" + productId + "/variants/" + variantId),
  createImage: (productId: string, input: { image_url: string; sort_order: number; is_primary: boolean }) =>
    api.post<ApiImage>("/products/" + productId + "/images", input),
  updateImage: (productId: string, imageId: string, input: Partial<{ sort_order: number; is_primary: boolean }>) =>
    api.patch<ApiImage>("/products/" + productId + "/images/" + imageId, input),
  deleteImage: (productId: string, imageId: string) => api.del("/products/" + productId + "/images/" + imageId),

  categories: () => api.list<ApiCategory>("/categories", { limit: 100, include_inactive: true }),
  createCategory: (input: { name: string; is_active: boolean }) => api.post<ApiCategory>("/categories", input),
  updateCategory: (id: string, input: Partial<{ name: string; is_active: boolean }>) =>
    api.patch<ApiCategory>("/categories/" + id, input),
  deleteCategory: (id: string) => api.del("/categories/" + id),

  banners: () => api.get<ApiBanner[]>("/banners", { include_inactive: true }),
  createBanner: (input: BannerInput) => api.post<ApiBanner>("/banners", input),
  updateBanner: (id: string, input: Partial<BannerInput>) => api.patch<ApiBanner>("/banners/" + id, input),
  deleteBanner: (id: string) => api.del("/banners/" + id),

  posts: (query?: { status?: string; q?: string; page?: number; limit?: number }) =>
    api.list<ApiPost>("/admin/posts", query),
  post: (id: string) => api.get<ApiPostDetail>("/admin/posts/" + id),
  createPost: (input: PostInput) => api.post<ApiPostDetail>("/admin/posts", input),
  updatePost: (id: string, input: Partial<PostInput>) => api.patch<ApiPostDetail>("/admin/posts/" + id, input),
  deletePost: (id: string) => api.del("/admin/posts/" + id),

  users: (query?: { q?: string; role?: string; status?: string; page?: number; limit?: number }) =>
    api.list<ApiUser>("/admin/users", query),
  updateUser: (id: string, input: Partial<{ role: string; status: string }>) =>
    api.patch<ApiUser>("/admin/users/" + id, input),

  updateSettings: (input: Partial<ApiSiteSettings>) => api.put<ApiSiteSettings>("/site-settings", input),

  /** Unggah gambar ke Cloudinary -> URL (JPG/PNG/WebP, maks 5 MB). */
  uploadImage: (file: File) => {
    const fd = new FormData();
    fd.append("file", file);
    return api.post<{ url: string }>("/admin/uploads/images", fd);
  },
};
