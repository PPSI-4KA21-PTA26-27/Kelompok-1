/** Tautan aman: path dalam situs ("/katalog", bukan "//host" atau "/\host") atau URL http(s). */
export const isSafeLink = (s: string): boolean => /^\/(?![/\\])|^https?:\/\//.test(s);
