/** OWNER / ADMIN may mutate; MEMBER is read-only. */
export function isEditor(role: string | undefined | null): boolean {
  return role === 'OWNER' || role === 'ADMIN';
}
