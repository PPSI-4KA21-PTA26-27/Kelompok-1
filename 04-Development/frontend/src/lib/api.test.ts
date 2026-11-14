import { describe, expect, it, vi } from "vitest";
import { api, API_URL, ApiError, errorMessage, fetchBlob, request, tokenStore } from "./api";
import { fetchCall, jsonResponse, mockFetch } from "@/test/fetch-mock";

describe("API_URL", () => {
  it("memangkas garis miring di akhir NEXT_PUBLIC_API_URL", () => {
    expect(API_URL).toBe("http://api.test/api/v1");
  });
});

describe("tokenStore", () => {
  it("menyimpan, membaca, dan menghapus token", () => {
    expect(tokenStore.get()).toBeNull();
    tokenStore.set("abc");
    expect(tokenStore.get()).toBe("abc");
    tokenStore.clear();
    expect(tokenStore.get()).toBeNull();
  });
});

describe("request", () => {
  it("menyusun URL dari API_URL + path dan membuang query kosong", async () => {
    const f = mockFetch();
    await request("/products", { query: { q: "kaos", page: 2, empty: "", nil: null, undef: undefined, flag: false } });
    const { url, method } = fetchCall(f);
    expect(url.origin + url.pathname).toBe("http://api.test/api/v1/products");
    expect(Object.fromEntries(url.searchParams)).toEqual({ q: "kaos", page: "2", flag: "false" });
    expect(method).toBe("GET");
  });

  it("mengirim Bearer JWT bila token ada", async () => {
    tokenStore.set("jwt-123");
    const f = mockFetch();
    await request("/users/me");
    expect(fetchCall(f).headers.Authorization).toBe("Bearer jwt-123");
    expect(fetchCall(f).headers.Accept).toBe("application/json");
  });

  it("tidak mengirim Authorization bila auth: false atau tanpa token", async () => {
    tokenStore.set("jwt-123");
    const f = mockFetch();
    await request("/auth/login", { auth: false });
    expect(fetchCall(f).headers.Authorization).toBeUndefined();
    tokenStore.clear();
    await request("/products");
    expect(fetchCall(f, 1).headers.Authorization).toBeUndefined();
  });

  it("mengirim body JSON dengan Content-Type", async () => {
    const f = mockFetch();
    await request("/x", { method: "POST", body: { a: 1 } });
    const c = fetchCall(f);
    expect(c.method).toBe("POST");
    expect(c.headers["Content-Type"]).toBe("application/json");
    expect(c.body).toEqual({ a: 1 });
  });

  it("FormData dikirim apa adanya tanpa Content-Type", async () => {
    const f = mockFetch();
    const fd = new FormData();
    fd.append("file", "x");
    await request("/upload", { method: "PUT", body: fd });
    const c = fetchCall(f);
    expect(c.body).toBe(fd);
    expect(c.headers["Content-Type"]).toBeUndefined();
  });

  it("membuka bungkus {data, meta}", async () => {
    mockFetch(() => jsonResponse({ data: [{ id: 1 }], meta: { page: 1, limit: 10, total: 1 } }));
    await expect(request("/x")).resolves.toEqual({ data: [{ id: 1 }], meta: { page: 1, limit: 10, total: 1 } });
  });

  it("204 menghasilkan data undefined", async () => {
    mockFetch(() => jsonResponse(null, 204));
    await expect(request("/x", { method: "DELETE" })).resolves.toEqual({ data: undefined });
  });

  it("respons gagal menjadi ApiError dengan code, message, details", async () => {
    mockFetch(() =>
      jsonResponse(
        { error: { code: "validation_error", message: "Input tidak valid", details: [{ field: "email", message: "Email wajib" }] } },
        422,
      ),
    );
    const err = await request("/x").catch((e) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect(err).toMatchObject({ status: 422, code: "validation_error", message: "Input tidak valid", name: "ApiError" });
    expect(err.details).toEqual([{ field: "email", message: "Email wajib" }]);
  });

  it("body error bukan JSON memakai kode dan pesan bawaan", async () => {
    mockFetch(() => new Response("oops", { status: 500 }));
    const err = await request("/x").catch((e) => e);
    expect(err).toMatchObject({ status: 500, code: "unknown_error", details: [] });
  });

  it("fetch gagal (jaringan) menjadi ApiError status 0", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("Failed to fetch")));
    const err = await request("/x").catch((e) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect(err).toMatchObject({ status: 0, code: "network_error" });
  });

  it("AbortError diteruskan apa adanya", async () => {
    const abort = new DOMException("aborted", "AbortError");
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(abort));
    await expect(request("/x")).rejects.toBe(abort);
  });

  it("401 dengan token: hapus token dan kirim event auth:expired", async () => {
    tokenStore.set("expired");
    mockFetch(() => jsonResponse({ error: { code: "unauthorized", message: "Token kedaluwarsa" } }, 401));
    const onExpired = vi.fn();
    window.addEventListener("auth:expired", onExpired);
    try {
      await expect(request("/users/me")).rejects.toMatchObject({ status: 401, code: "unauthorized" });
      expect(tokenStore.get()).toBeNull();
      expect(onExpired).toHaveBeenCalledTimes(1);
    } finally {
      window.removeEventListener("auth:expired", onExpired);
    }
  });

  it("401 tanpa token (mis. login gagal) tidak memicu auth:expired", async () => {
    mockFetch(() => jsonResponse({ error: { code: "invalid_credentials", message: "Salah" } }, 401));
    const onExpired = vi.fn();
    window.addEventListener("auth:expired", onExpired);
    try {
      await expect(request("/auth/login", { auth: false })).rejects.toBeInstanceOf(ApiError);
      expect(onExpired).not.toHaveBeenCalled();
    } finally {
      window.removeEventListener("auth:expired", onExpired);
    }
  });
});

describe("api helpers", () => {
  it("get mengembalikan data saja", async () => {
    mockFetch(() => jsonResponse({ data: { id: "u1" } }));
    await expect(api.get("/users/me")).resolves.toEqual({ id: "u1" });
  });

  it("list memberi default data [] dan meta bila kosong", async () => {
    mockFetch(() => jsonResponse({}));
    await expect(api.list("/products")).resolves.toEqual({ data: [], meta: { page: 1, limit: 0, total: 0 } });
  });

  it.each([
    ["post", "POST"],
    ["put", "PUT"],
    ["patch", "PATCH"],
  ] as const)("%s memakai method %s", async (fn, method) => {
    const f = mockFetch();
    await api[fn]("/x", { a: 1 });
    expect(fetchCall(f)).toMatchObject({ method, body: { a: 1 } });
  });

  it("del memakai DELETE", async () => {
    const f = mockFetch(() => jsonResponse(null, 204));
    await api.del("/x");
    expect(fetchCall(f).method).toBe("DELETE");
  });
});

describe("fetchBlob", () => {
  it("mengirim token dan mengembalikan Blob", async () => {
    tokenStore.set("t");
    const f = mockFetch(() => new Response("isi", { status: 200 }));
    const blob = await fetchBlob("/orders/o1/payment/proof");
    expect(await blob.text()).toBe("isi");
    expect(fetchCall(f).headers.Authorization).toBe("Bearer t");
  });

  it("gagal menjadi ApiError", async () => {
    mockFetch(() => jsonResponse({ error: { code: "not_found", message: "Tidak ada" } }, 404));
    await expect(fetchBlob("/x")).rejects.toMatchObject({ status: 404, code: "not_found" });
  });
});

describe("errorMessage", () => {
  it("menggabungkan details ApiError bila ada", () => {
    const e = new ApiError(422, "v", "umum", [
      { field: "a", message: "A salah." },
      { field: "b", message: "B salah." },
    ]);
    expect(errorMessage(e)).toBe("A salah. B salah.");
  });

  it("memakai message ApiError / Error, atau pesan bawaan", () => {
    expect(errorMessage(new ApiError(500, "x", "Server error"))).toBe("Server error");
    expect(errorMessage(new Error("biasa"))).toBe("biasa");
    expect(errorMessage("string")).toBe("Terjadi kesalahan.");
  });
});
