import { vi } from "vitest";

/** Respons JSON palsu untuk fetch. */
export function jsonResponse(body: unknown, status = 200): Response {
  if (status === 204) return new Response(null, { status });
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

/** Ganti global fetch dengan mock yang selalu mengembalikan respons dari factory. */
export function mockFetch(factory: () => Response = () => jsonResponse({ data: null })) {
  const fn = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>(async () => factory());
  vi.stubGlobal("fetch", fn);
  return fn;
}

/** Ambil URL, method, header, dan body (JSON di-parse) dari panggilan fetch ke-n. */
export function fetchCall(fn: ReturnType<typeof mockFetch>, n = 0) {
  const [input, init = {}] = fn.mock.calls[n];
  const url = new URL(String(input));
  const headers = (init.headers ?? {}) as Record<string, string>;
  const raw = init.body;
  const body = typeof raw === "string" ? JSON.parse(raw) : raw;
  return { url, path: url.pathname, method: init.method ?? "GET", headers, body };
}
