# Architecture Decision Records

## ADR-0001 — Broker port takes a batch, returns per-order results (2026-07-02)
`PlaceOrder(ctx, []Order) []OrderResult`. Ports are consumer-defined: the engine places a set of orders per tick, so batching is its vocabulary — whether a broker has a native batch API is the adapter's problem (Kite has none; it fans out with a throttled worker pool). Outcomes are independent (fill / reject / timeout), so results are per-order, each keyed by our own `ClientOrderID`.
