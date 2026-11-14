import { beforeEach, describe, expect, it } from "vitest";
import { adminService, authService, cartService, orderService, paymentService } from "./index";
import { tokenStore } from "@/lib/api";
import { fetchCall, mockFetch } from "@/test/fetch-mock";

const BASE = "/api/v1";
let f: ReturnType<typeof mockFetch>;

beforeEach(() => {
  tokenStore.set("jwt");
  f = mockFetch();
});

const call = (n = 0) => {
  const c = fetchCall(f, n);
  return { path: c.path.replace(BASE, ""), method: c.method, body: c.body, auth: c.headers.Authorization, query: c.url.searchParams };
};

describe("authService", () => {
  it("login: POST /auth/login tanpa Authorization", async () => {
    await authService.login("a@b.c", "rahasia");
    expect(call()).toMatchObject({ path: "/auth/login", method: "POST", body: { email: "a@b.c", password: "rahasia" }, auth: undefined });
  });

  it("register: POST /auth/register tanpa Authorization", async () => {
    const input = { name: "A", email: "a@b.c", password: "x", phone: null };
    await authService.register(input);
    expect(call()).toMatchObject({ path: "/auth/register", method: "POST", body: input, auth: undefined });
  });

  it("me: GET /users/me dengan Bearer", async () => {
    await authService.me();
    expect(call()).toMatchObject({ path: "/users/me", method: "GET", auth: "Bearer jwt" });
  });

  it("changePassword: PUT /users/me/password", async () => {
    await authService.changePassword({ current_password: "a", new_password: "b" });
    expect(call()).toMatchObject({ path: "/users/me/password", method: "PUT", body: { current_password: "a", new_password: "b" } });
  });
});

describe("cartService", () => {
  it("get: GET /cart", async () => {
    await cartService.get();
    expect(call()).toMatchObject({ path: "/cart", method: "GET", auth: "Bearer jwt" });
  });

  it("add: POST /cart/items", async () => {
    await cartService.add("v1", 2);
    expect(call()).toMatchObject({ path: "/cart/items", method: "POST", body: { variant_id: "v1", quantity: 2 } });
  });

  it("update: PATCH /cart/items/:id", async () => {
    await cartService.update("i1", 3);
    expect(call()).toMatchObject({ path: "/cart/items/i1", method: "PATCH", body: { quantity: 3 } });
  });

  it("remove dan clear memakai DELETE", async () => {
    await cartService.remove("i1");
    await cartService.clear();
    expect(call(0)).toMatchObject({ path: "/cart/items/i1", method: "DELETE" });
    expect(call(1)).toMatchObject({ path: "/cart", method: "DELETE" });
  });
});

describe("orderService", () => {
  it("listMine: GET /orders dengan query", async () => {
    await orderService.listMine({ page: 2, order_status: "paid" });
    const c = call();
    expect(c).toMatchObject({ path: "/orders", method: "GET" });
    expect(c.query.get("page")).toBe("2");
    expect(c.query.get("order_status")).toBe("paid");
  });

  it("create: POST /orders dengan input checkout", async () => {
    const input = {
      recipient_name: "A",
      recipient_phone: "08",
      shipping_address: "Jl. X",
      province_id: 1,
      city_id: 2,
      district_id: 3,
      subdistrict_id: 4,
      cart_item_ids: ["c1"],
    };
    await orderService.create(input);
    expect(call()).toMatchObject({ path: "/orders", method: "POST", body: input });
  });

  it("cancel: body reason hanya bila diisi", async () => {
    await orderService.cancel("o1", "salah pesan");
    await orderService.cancel("o2");
    expect(call(0)).toMatchObject({ path: "/orders/o1/cancel", method: "POST", body: { reason: "salah pesan" } });
    expect(call(1)).toMatchObject({ path: "/orders/o2/cancel", method: "POST", body: undefined });
  });

  it("complete: POST /orders/:id/complete", async () => {
    await orderService.complete("o1");
    expect(call()).toMatchObject({ path: "/orders/o1/complete", method: "POST" });
  });
});

describe("paymentService", () => {
  it("create: POST /orders/:id/payment metode qris", async () => {
    await paymentService.create("o1");
    expect(call()).toMatchObject({ path: "/orders/o1/payment", method: "POST", body: { method: "qris" } });
  });

  it("uploadProof: PUT multipart dengan field file", async () => {
    const file = new File(["x"], "bukti.png", { type: "image/png" });
    await paymentService.uploadProof("o1", file);
    const c = call();
    expect(c).toMatchObject({ path: "/orders/o1/payment/proof", method: "PUT" });
    expect(c.body).toBeInstanceOf(FormData);
    expect((c.body as FormData).get("file")).toBeInstanceOf(File);
  });
});

describe("adminService", () => {
  it("upsertShipment: PUT /orders/:id/shipment", async () => {
    const input = { tracking_number: null, coverage_status: "covered", shipping_status: "processing" };
    await adminService.upsertShipment("o1", input);
    expect(call()).toMatchObject({ path: "/orders/o1/shipment", method: "PUT", body: input, auth: "Bearer jwt" });
  });

  it("reviewPayment: POST /admin/payments/:id/review", async () => {
    await adminService.reviewPayment("p1", "approved");
    expect(call()).toMatchObject({ path: "/admin/payments/p1/review", method: "POST", body: { status: "approved" } });
  });

  it("updateOrderStatus: PUT /admin/orders/:id/status", async () => {
    await adminService.updateOrderStatus("o1", { order_status: "shipped", tracking_number: "R1" });
    expect(call()).toMatchObject({ path: "/admin/orders/o1/status", method: "PUT", body: { order_status: "shipped", tracking_number: "R1" } });
  });
});
