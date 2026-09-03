/** Presentation helpers for contract revision actions. Canonical status comes from API DTOs. */

export function canValidateRevision(status: string | undefined | null): boolean {
  return status === 'DRAFT';
}

export function canActivateRevision(status: string | undefined | null): boolean {
  return status === 'VALIDATED';
}

export function canDeprecateRevision(status: string | undefined | null): boolean {
  return status === 'ACTIVE' || status === 'STALE';
}
