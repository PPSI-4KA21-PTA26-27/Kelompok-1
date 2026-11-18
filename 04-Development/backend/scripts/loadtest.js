// Uji beban endpoint publik (read-only). Jalankan dari mesin lain, bukan server:
//   k6 run -e BASE=https://<app>.up.railway.app/api/v1 backend/scripts/loadtest.js
// Pakai 20 VU = perkiraan puncak pengguna bersamaan. Naikkan bertahap bila perlu.
import http from "k6/http";
import { check, sleep } from "k6";

const BASE = __ENV.BASE;

export const options = {
  stages: [
    { duration: "30s", target: 20 },
    { duration: "2m", target: 20 },
    { duration: "15s", target: 0 },
  ],
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<500", "p(99)<1000"],
  },
};

const paths = ["/products", "/categories", "/banners", "/site-settings", "/health"];

export default function () {
  const res = http.get(BASE + paths[Math.floor(Math.random() * paths.length)], {
    headers: { "Accept-Encoding": "gzip" },
  });
  check(res, { "status 200": (r) => r.status === 200 });
  sleep(1);
}
