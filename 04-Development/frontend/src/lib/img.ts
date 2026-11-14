/**
 * Optimises Cloudinary image URLs on the fly (auto format, auto quality,
 * width cap) so mobile devices don't download full-size originals.
 * Non-Cloudinary URLs and URLs that already carry a transformation are
 * returned unchanged.
 */
export function cdn(url: string | null | undefined, w: number): string {
  if (!url) return "";
  if (!url.includes("res.cloudinary.com")) return url;
  const marker = "/upload/";
  const i = url.indexOf(marker);
  if (i < 0) return url;
  const rest = url.slice(i + marker.length);
  // already transformed (e.g. "f_auto,..." or "w_300/...")
  if (/^[a-z]{1,3}_[^/]+\//.test(rest)) return url;
  return (
    url.slice(0, i + marker.length) +
    `f_auto,q_auto,c_limit,w_${Math.round(w)}/` +
    rest
  );
}
