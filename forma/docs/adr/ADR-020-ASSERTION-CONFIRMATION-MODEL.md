# ADR-020: Assertion Confirmation Model

## Status

Accepted (FORMA S3)

## Decision

Assertions have lifecycle status (`PROPOSED`, `CONFIRMED`, `REJECTED`, `SUPERSEDED`). Confirm/Reject creates immutable `BusinessConfirmation` events. High model confidence does not auto-confirm.

## Consequences

- Repeat Confirm/Reject returns `FORMA_ASSERTION_ALREADY_DECIDED`.
- Only CONFIRMED assertions enter Proposal builder by default.
