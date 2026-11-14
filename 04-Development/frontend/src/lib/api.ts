import type { ApiMeta, Paginated } from "@/types/api";

export const API_URL = (process.env.NEXT_PUBLIC_API_URL as string).replace(
  /\/$/,
  "",
);

const TOKEN_KEY = "hs_access_token";

/** Token JWT disimpan di localStorage (backend memakai header Authorization, bukan cookie). */
export const tokenStore = {
  get(): string | null {
    try {
      return localStorage.getItem(TOKEN_KEY);
    } catch {
      return null;
    }
  },
  set(token: string) {
    try {
      localStorage.setItem(TOKEN_KEY, token);
    } catch {}
  },
  clear() {
    try {
      localStorage.removeItem(TOKEN_KEY);
    } catch {}
  },
};

export class ApiError extends Error {
  status: number;
  code: string;
  details: { field: string; message: string }[];
  constructor(
    status: number,
    code: string,
    message: string,
    details: { field: string; message: string }[] = [],
  ) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

type Query = Record<string, string | number | boolean | null | undefined>;
type Options = {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  body?: unknown;
  query?: Query;
  /** Kirim header Authorization bila token ada (default true). */
  auth?: boolean;
  signal?: AbortSignal;
};

function buildUrl(path: string, query?: Query) {
  const url = new URL(API_URL + path);
  if (query) {
    for (const [k, v] of Object.entries(query)) {
      if (v !== undefined && v !== null && v !== "")
        url.searchParams.set(k, String(v));
    }
  }
  return url.toString();
}

export async function request<T>(
  path: string,
  { method = "GET", body, query, auth = true, signal }: Options = {},
): Promise<{ data: T; meta?: ApiMeta }> {
  const headers: Record<string, string> = { Accept: "application/json" };
  const token = auth ? tokenStore.get() : null;
  if (token) headers.Authorization = "Bearer " + token;

  let payload: BodyInit | undefined;
  if (body instanceof FormData) {
    payload = body; // Content-Type (multipart boundary) diisi browser
  } else if (body !== undefined) {
    headers["Content-Type"] = "application/json";
    payload = JSON.stringify(body);
  }

  let res: Response;
  try {
    res = await fetch(buildUrl(path, query), {
      method,
      headers,
      body: payload,
      signal,
    });
  } catch (e) {
    if ((e as Error).name === "AbortError") throw e;
    throw new ApiError(
      0,
      "network_error",
      "Tidak dapat terhubung ke server. Pastikan backend berjalan.",
    );
  }

  if (res.status === 204) return { data: undefined as T };

  const json = await res.json().catch(() => null);
  if (!res.ok) {
    if (res.status === 401 && token && typeof window !== "undefined") {
      tokenStore.clear();
      window.dispatchEvent(new Event("auth:expired"));
    }
    const err = json?.error;
    throw new ApiError(
      res.status,
      err?.code ?? "unknown_error",
      err?.message ?? "Terjadi kesalahan pada server.",
      err?.details ?? [],
    );
  }
  return { data: json?.data as T, meta: json?.meta };
}

export const api = {
  get: <T>(path: string, query?: Query, opts?: Options) =>
    request<T>(path, { ...opts, query }).then((r) => r.data),
  list: <T>(path: string, query?: Query, opts?: Options) =>
    request<T[]>(path, { ...opts, query }).then(
      (r) =>
        ({
          data: r.data ?? [],
          meta: r.meta ?? { page: 1, limit: 0, total: 0 },
        }) as Paginated<T>,
    ),
  post: <T>(path: string, body?: unknown, opts?: Options) =>
    request<T>(path, { ...opts, method: "POST", body }).then((r) => r.data),
  put: <T>(path: string, body?: unknown, opts?: Options) =>
    request<T>(path, { ...opts, method: "PUT", body }).then((r) => r.data),
  patch: <T>(path: string, body?: unknown, opts?: Options) =>
    request<T>(path, { ...opts, method: "PATCH", body }).then((r) => r.data),
  del: <T = void>(path: string, opts?: Options) =>
    request<T>(path, { ...opts, method: "DELETE" }).then((r) => r.data),
};

/** Ambil berkas biner yang butuh token (mis. bukti bayar) sebagai Blob. */
export async function fetchBlob(path: string): Promise<Blob> {
  const headers: Record<string, string> = {};
  const token = tokenStore.get();
  if (token) headers.Authorization = "Bearer " + token;
  const res = await fetch(buildUrl(path), { headers });
  if (!res.ok) {
    const json = await res.json().catch(() => null);
    throw new ApiError(
      res.status,
      json?.error?.code ?? "unknown_error",
      json?.error?.message ?? "Gagal memuat berkas.",
    );
  }
  return res.blob();
}

/** Pesan error ramah-pengguna dari ApiError / Error biasa. */
export function errorMessage(e: unknown): string {
  if (e instanceof ApiError) {
    if (e.details.length) return e.details.map((d) => d.message).join(" ");
    return e.message;
  }
  return e instanceof Error ? e.message : "Terjadi kesalahan.";
}
