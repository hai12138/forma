/**
 * Safe internal returnTo for Forma login redirects.
 * Rejects open redirects (protocol-relative, absolute URLs, empty).
 */
export function safeReturnTo(raw: string | null | undefined, fallback = '/'): string {
  if (!raw) return fallback;
  let decoded = raw;
  try {
    decoded = decodeURIComponent(raw);
  } catch {
    return fallback;
  }
  if (!decoded.startsWith('/')) return fallback;
  if (decoded.startsWith('//')) return fallback;
  if (decoded.includes('://')) return fallback;
  if (decoded.includes('\\')) return fallback;
  return decoded;
}

export function encodeReturnTo(path: string): string {
  return encodeURIComponent(safeReturnTo(path));
}
