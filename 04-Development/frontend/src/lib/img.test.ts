import { describe, expect, it } from "vitest";
import { cdn } from "./img";

const base = "https://res.cloudinary.com/demo/image/upload/";

describe("cdn", () => {
  it("menambah transformasi pada URL Cloudinary", () => {
    expect(cdn(base + "a.jpg", 300)).toBe(base + "f_auto,q_auto,c_limit,w_300/a.jpg");
  });
  it("tidak mengubah URL yang sudah punya transformasi", () => {
    expect(cdn(base + "w_100/a.jpg", 300)).toBe(base + "w_100/a.jpg");
  });
  it("tidak mengubah host lain yang memuat res.cloudinary.com", () => {
    const a = "https://evil.com/res.cloudinary.com/upload/a.jpg";
    const b = "https://res.cloudinary.com.evil.com/upload/a.jpg";
    expect(cdn(a, 300)).toBe(a);
    expect(cdn(b, 300)).toBe(b);
  });
  it("aman untuk kosong dan URL relatif", () => {
    expect(cdn(null, 300)).toBe("");
    expect(cdn("/lokal.jpg", 300)).toBe("/lokal.jpg");
  });
});
