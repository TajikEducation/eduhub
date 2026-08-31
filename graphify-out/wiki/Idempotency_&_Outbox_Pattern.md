# Idempotency & Outbox Pattern

> 4 nodes · cohesion 0.50

## Key Concepts

- **Inter-Service Interaction (Interfaces + Outbox)** (2 connections) — `docs/EduHub_Backend_Architecture.md`
- **reviews.outbox** (2 connections) — `docs/EduHub_Database_Schema.md`
- **communications.processed_events** (1 connections) — `docs/EduHub_Database_Schema.md`
- **platform.idempotency_keys** (1 connections) — `docs/EduHub_Database_Schema.md`

## Relationships

- [[Redis Cache & Contract Risks]] (6 shared connections)

## Source Files

- `docs/EduHub_Backend_Architecture.md`
- `docs/EduHub_Database_Schema.md`

## Audit Trail

- EXTRACTED: 4 (67%)
- INFERRED: 2 (33%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*