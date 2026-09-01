# ADR-018: AI Does Not Directly Mutate Business Model

## Status

Accepted (FORMA S3)

## Context

AI-generated business facts must be traceable and human-approved before entering the Business Model SoT.

## Decision

AI Analyst operations may create Evidence, Assertions, and Proposals. Only `ApplyProposal` (human-authorized, server-side) creates a new Business Model Revision via S2 `BusinessService.SaveModel`.

## Consequences

- Interview alone never changes `current_revision`.
- All AI elements carry `AI_GENERATED` source markers until human edit.
