# Architecture Decision Records

> Convention: each entry is max two lines — the decision, then the why.

## ADR-0001 — Broker port takes a batch, returns per-order results (2026-07-02)
`PlaceOrder(ctx, []Order) []OrderResult`. Ports are consumer-defined: the engine places a set of orders per tick, so batching is its vocabulary — whether a broker has a native batch API is the adapter's problem (Kite has none; it fans out with a throttled worker pool). Outcomes are independent (fill / reject / timeout), so results are per-order, each keyed by our own `ClientOrderID`.

## ADR-0002 — Strategies own their instrument universe (2026-07-05)
`Strategy` exposes `Instruments() []Instrument`; the engine subscribes to what the strategy declares rather than to an external list. The strategy inherently knows what it trades, so the need flows from it (no list drift), and the same method seeds the Phase-9 multi-strategy subscription union and routing.
