# Algo Trading System in Go — Roadmap & Review Reference

> A learning-first roadmap to build a broker-agnostic algorithmic trading system in Go.
> This document is both a **build guide** ("what to do next, and what I don't yet know")
> and a **review rubric** (the bar each phase is held to). No code or folder structure is
> prescribed here by design — you build by hand; this tells you *what* to build and *why*.

---

## Context

**Why this exists.** The goal is to *learn* — production-grade Go, clean architecture, and the
real mechanics of algorithmic trading — not to make money. Success is measured by code quality,
extensibility, and how much of the domain you genuinely understand, not by P&L.

**Scope (decided up front):**
- **Execution modes:** Backtesting + Paper trading + Live trading — all three.
- **Market data:** Both WebSocket streaming (real-time ticks) *and* REST (historical/snapshots).
- **Strategy:** A pluggable strategy framework, seeded with a few simple technical-analysis
  strategies (e.g. SMA/EMA crossover, RSI).
- **Ops:** Full production stack — config, structured logging, metrics/observability, testing,
  containerization, CI/CD.
- **Broker:** Only Zerodha (Kite) will be integrated, but **everything broker-specific lives
  behind interfaces** so a second broker could be added without touching strategy or engine code.

**Non-goals:** real profit, high-frequency/low-latency optimization, options greeks/derivatives
pricing, ML strategies, a polished web UI. Any of these are optional stretch goals later.

---

## The One Idea That Makes Everything Else Work: an Event-Driven Core

This is the architectural spine. Internalize it before writing anything.

A trading system is a pipeline of **events** flowing through components:

```
DATA SOURCE → [MarketEvent] → STRATEGY → [SignalEvent] → PORTFOLIO/RISK → [OrderEvent] → BROKER → [FillEvent] → PORTFOLIO
     ▲                                                                                                              │
     └──────────────────────────────── (loop; portfolio state informs next decisions) ◄──────────────────────────┘
```

- **MarketEvent** — new data arrived (a tick or a completed candle).
- **SignalEvent** — the strategy's intent ("I want to go long X"), *not* an order.
- **OrderEvent** — a concrete, sized, risk-checked order ready to send.
- **FillEvent** — the broker reported an execution (full or partial), which updates positions/cash.

**Why this is the whole game:** the *only* things that differ between backtest, paper, and live
are (a) where MarketEvents come from and (b) what handles OrderEvents and produces FillEvents.
Strategy code, portfolio accounting, and risk checks are **identical in all three modes**.

> **Litmus test you should apply forever:** if you ever write `if backtest { ... } else { ... }`
> inside a strategy, indicator, portfolio, or risk component, your abstraction has leaked. The
> mode should be invisible to everything above the "ports."

This is **hexagonal architecture (ports & adapters)**: the core domain depends only on interfaces
("ports"); Kite, a CSV reader, a simulated broker, Postgres, etc. are interchangeable "adapters."

---

## Part A — Disciplines to Apply From Day One

These are *not* a phase. They are habits that must be present in Phase 0's first commit and every
commit after. They are also the first thing reviewed.

### A1. Go idioms & clean code
- Accept interfaces, return structs. Keep interfaces **small** and defined by the *consumer*, not
  the implementer. A broker interface with 30 methods is a smell.
- Packages organized by **domain capability**, not by technical layer ("models/", "utils/" are
  anti-patterns). Learn the `cmd/` vs `internal/` vs `pkg/` conventions — then decide your own
  layout deliberately.
- No global mutable state. Dependencies are passed in (constructor injection), not reached for.
- `gofmt`/`goimports` always; `go vet` and `golangci-lint` clean before every commit.

### A2. Domain modeling correctness (two traps that bite everyone)
- **Money is never a `float64`.** `0.1 + 0.2 != 0.3`. Represent prices/cash as integer minor units
  (paise) or a decimal type (`shopspring/decimal`). Decide once, early, and wrap it in a type.
- **Time is timezone- and calendar-aware.** Markets run in IST, have session windows, and have
  holidays. Model a market clock / session concept; never assume "now" maps to "market open."

### A3. Error handling
- Wrap errors with context (`fmt.Errorf("...: %w", err)`); define sentinel/typed errors for
  conditions callers must branch on (e.g. "order rejected", "insufficient margin").
- Decide deliberately what is *recoverable* (retry) vs *fatal* (halt and alert). In a live trading
  loop this distinction is safety-critical.

### A4. Concurrency & lifecycle
- A live system is inherently concurrent: a goroutine reading the WebSocket, the strategy loop,
  order submission, reconciliation. Coordinate with **channels + `context.Context`**, not shared
  locks where avoidable.
- **Graceful shutdown** on SIGINT/SIGTERM: stop accepting new signals, let in-flight orders settle,
  flush state, close connections. A trading bot that dies mid-order is dangerous.
- Run tests and the bot with the **race detector** (`-race`) routinely.

### A5. Configuration & secrets
- 12-factor style: config via env / files, never hard-coded. API key/secret/access-token **never**
  in source or git. Use `.env` (git-ignored) locally; a secrets mechanism in deployment.
- One typed config struct, validated at startup; fail fast on missing/invalid config.

### A6. Testing strategy
- **Table-driven unit tests** for indicators, sizing, risk rules, accounting — these are pure
  functions and must be bulletproof.
- Interfaces make mocking trivial; generate mocks from your ports (e.g. `mockgen`/`moq`).
- Your **backtester is itself a giant integration test** for strategies + the event engine.
- Aim for meaningful coverage on the *domain core*; don't chase 100% on glue code.

### A7. Observability baseline
- Structured logging via the stdlib `log/slog` from day one — leveled, with fields (instrument,
  order id, strategy). No `fmt.Println` debugging that ships.
- Plan for metrics (Prometheus) and health endpoints; wire them properly in Phase 8 but don't
  retrofit logging at the end.

### A8. Repo hygiene
- Conventional commits, a real README, an ADR (architecture decision record) log for "why" choices.
- `Makefile`/`Taskfile` for build/test/lint/run so the project is reproducible.

---

## Part B — Phased Build Roadmap

Each phase lists: **Goal**, **Build (concepts)**, **Key abstractions**, **Done when**, **Traps**.
Build in order — each phase de-risks the next. You'll have something runnable and reviewable by the
end of Phase 2, well before any broker integration.

### Phase 0 — Foundations & domain model
- **Goal:** A typed vocabulary of the domain and the project skeleton with disciplines (Part A) in place.
- **Build:** Core value types — Instrument/Symbol, Price/Money, Quantity, Side (buy/sell),
  Tick, Candle/Bar (OHLCV), Order (type, product, validity, qty, price), Position, Trade/Fill,
  Portfolio (cash + positions). A market-clock/session concept.
- **Done when:** these types compile, are unit-tested, and money/time are handled correctly (A2).
- **Traps:** premature abstraction — model only what you understand now; floats for money.

### Phase 1 — Core ports (interfaces) & the event engine
- **Goal:** Define the seams that make the system swappable, and the loop that moves events.
- **Build (as interfaces, by responsibility — not implementations yet):**
  - **MarketDataProvider** — subscribe to instruments; deliver MarketEvents (historical replay *or*
    live stream sit behind the same port).
  - **Strategy** — consume MarketEvents, emit SignalEvents; holds its own indicator state.
  - **Portfolio** — apply FillEvents, track positions/cash/P&L, answer "what do I hold?".
  - **RiskManager** — vet a proposed order; approve/resize/reject.
  - **Broker / ExecutionHandler** — accept OrderEvents, produce FillEvents; query orders/positions.
  - **Repository/Store** — persist orders, fills, candles, audit log.
  - The **engine/runner** that wires these and pumps events through.
- **Done when:** the interfaces exist with zero concrete implementations and the engine compiles
  against them. This forces you to design boundaries before being seduced by Kite's API shape.
- **Traps:** letting Kite's data structures leak into your core types (define *your* domain model;
  adapters translate to/from it); interfaces that are too fat.

### Phase 2 — Backtesting engine (offline, no broker, no network)
- **Goal:** Prove the architecture end-to-end with a deterministic, fully testable simulation.
- **Build:** A historical MarketDataProvider (replays candles from CSV/DB), a **SimulatedBroker**
  that fills OrderEvents against historical data, a commission/brokerage model, a slippage model,
  and a performance/metrics module (total return, CAGR, Sharpe, max drawdown, win rate, exposure).
- **Done when:** you can run a trivial strategy over historical data and get a metrics report,
  reproducibly (same input → same output).
- **Traps (the big ones):**
  - **Lookahead bias** — a strategy must not act on information it couldn't have had in real time
    (e.g. don't decide on a bar's close and fill at that same close; fill at the *next* bar's open).
  - **Survivorship bias** — testing only on instruments that still exist flatters results.
  - Indicator **warm-up** — don't trade before an indicator has enough data.
  - Forgetting costs — frictionless backtests lie; model commission + slippage from the start.

### Phase 3 — Strategy framework & indicators
- **Goal:** Make strategies genuinely pluggable and ship 2–3 real ones.
- **Build:** A clean Strategy contract + registration mechanism; reusable, tested indicators
  (SMA, EMA, RSI, etc.) as composable pieces; concrete strategies (e.g. SMA crossover, RSI
  mean-reversion). Parameterize strategies via config so the same code runs different settings.
- **Done when:** adding a new strategy requires implementing one interface and registering it —
  no changes to the engine, broker, or data layers. Each strategy validated via the backtester.
- **Traps:** strategies holding state incorrectly across events; over-fitting parameters to history.

### Phase 4 — Kite market-data adapter (REST + WebSocket) + auth
- **Goal:** First real Zerodha integration — *read-only* market data. No order placement yet.
- **Build:**
  - A **KiteDataProvider** adapter implementing MarketDataProvider using the official Go SDK
    (`github.com/zerodha/gokiteconnect/v4`): historical candles via REST, live ticks via the
    Ticker WebSocket (with auto-reconnect + resubscribe).
  - The **auth flow**: API key/secret → login URL → `request_token` → exchange for `access_token`.
  - A **rate-limiting** layer respecting Kite's limits (historical 3 req/s, quote 1 req/s).
  - Translation between Kite's payloads and *your* domain types (the adapter's job — keep Kite types
    out of the core).
- **Done when:** you can stream live ticks and pull historical candles through the *same*
  MarketDataProvider port your backtester uses.
- **Traps:**
  - **`access_token` expires daily** — there is no permanent token; design for a daily re-login
    (this is a genuine automation headache; plan how the bot obtains a fresh token each morning).
  - WebSocket realities: reconnection, resubscription after reconnect, heartbeat/staleness
    detection, and **backpressure** (what happens when ticks arrive faster than you process them).

### Phase 5 — Paper trading (live data + simulated execution)
- **Goal:** Run strategies against *live* market data with *simulated* fills and real-time portfolio
  tracking — the safe rehearsal for live.
- **Build:** Compose the live KiteDataProvider (Phase 4) with the SimulatedBroker (Phase 2).
  Because Zerodha has **no sandbox**, this in-house simulator *is* your paper-trading environment.
  Add real-time P&L tracking and session logging.
- **Done when:** the bot trades a strategy live (paper) for a full session and you can audit every
  signal → order → simulated fill.
- **Traps:** simulating fills naively in live conditions (a market order won't always fill at the
  last tick); divergence between paper assumptions and live reality (note these for later).

### Phase 6 — Live trading (real Kite broker adapter + order management)
- **Goal:** Place real orders, safely. This is where correctness matters most.
- **Build:**
  - A **KiteBroker** adapter implementing the Broker port: place/modify/cancel orders, query order
    status, positions, holdings, margins. Map your domain Order → Kite order varieties (regular,
    AMO, cover, iceberg), product types (CNC/MIS/NRML), and validity (DAY/IOC/TTL).
  - An **Order Management System (OMS)**: an explicit order **state machine**
    (NEW → PENDING → OPEN → PARTIALLY_FILLED → COMPLETE / REJECTED / CANCELLED) with **idempotency**
    (never double-send), and retry policy for transient failures.
  - **Reconciliation:** consume Kite **postbacks/webhooks** (COMPLETE/CANCEL/REJECTED/UPDATE) as the
    primary order-update channel, with REST polling as a fallback. Reconcile broker truth vs your
    internal state continuously.
  - **Crash recovery:** on restart, rebuild state from the broker (open orders/positions) + your
    store — never assume your in-memory view survived.
- **Done when:** the bot places, tracks, and reconciles a real order end-to-end, and recovers
  correct state after a forced restart.
- **Traps:**
  - **Partial fills** are normal — your accounting must handle them.
  - **Idempotency / exactly-once** — network timeouts make "did my order go through?" a real
    question; design so a retry can't create a duplicate order.
  - **Order rate limits** (10/s, 400/min) — throttle and queue.
  - **Static IP requirement** — since 1 Apr 2025, placing orders via the API requires a registered
    static IP. This is a *deployment* constraint (a fixed egress IP), surface it now.

### Phase 7 — Risk management & safety
- **Goal:** Make it hard for the system to hurt you.
- **Build:** Pre-trade risk checks (max position size, max order value, max open exposure,
  per-strategy capital allocation), daily max-loss limit, and a **kill switch** (flatten everything
  / halt new orders on a trigger or manual command). Risk runs *before* every OrderEvent reaches the
  broker.
- **Done when:** a strategy that tries to breach a limit is blocked, logged, and alerted — and the
  kill switch demonstrably halts trading.
- **Traps:** risk checks that can be bypassed by a code path; limits expressed in floats; no
  alerting when a limit trips.

### Phase 8 — Production hardening
- **Goal:** Make it operable, observable, and deployable like real software.
- **Build:**
  - **Persistence:** a real store (SQLite to start, or Postgres) for orders, fills, candles, and an
    immutable **audit log**. Migrations.
  - **Observability:** Prometheus metrics (orders placed, fills, latency, P&L, errors), `slog`
    structured logs shipped somewhere, health/readiness endpoints; consider OpenTelemetry tracing.
  - **Config & secrets** finalized; **graceful shutdown** verified.
  - **Containerization:** multi-stage Docker build (small, non-root image).
  - **CI/CD:** GitHub Actions running `go test -race`, `go vet`, `golangci-lint`, build, and image
    publish. Branch protection.
  - **Deployment:** somewhere with a static IP (Kite requirement), process supervision, restart
    policy, and a documented daily-token-refresh procedure.
- **Done when:** `docker run` (with config) boots the bot, metrics/logs are visible, CI is green,
  and there's a written runbook.
- **Traps:** retrofitting logging/metrics at the end; secrets baked into images; no restart strategy.

### Phase 9 — Stretch goals (pick what teaches you most)
- A CLI or small read-only dashboard (TUI/web) for live status & P&L.
- Notifications (Telegram/Slack/email) for fills, errors, kill-switch.
- Multi-strategy / portfolio-level allocation and netting.
- A second broker adapter (even a stub) to *prove* the abstraction holds — the ultimate test of the
  whole design.
- Walk-forward / out-of-sample backtest analysis; parameter optimization.

---

## Part C — "What You Don't Know You Don't Know" (consolidated trap list)

Keep this list visible; most are invisible until they hurt.

1. **Floats for money** — silent rounding corruption.
2. **Lookahead & survivorship bias** — backtests that look great and mean nothing.
3. **Frictionless backtests** — ignoring commission/slippage inverts conclusions.
4. **Daily access-token expiry** — no permanent auth; needs a daily refresh story.
5. **Static-IP requirement** for order placement (since Apr 2025) — a deployment constraint.
6. **Rate limits** (orders 10/s & 400/min; quote 1/s; historical 3/s) — must throttle.
7. **Partial fills** — the default assumption "fill = full" is wrong.
8. **Idempotency** — timeouts make duplicate orders a real risk.
9. **Reconciliation** — your view and the broker's view *will* diverge; design for it.
10. **Crash recovery** — rebuild from broker + store on restart; never trust in-memory survival.
11. **WebSocket reconnection, resubscription, staleness, and backpressure.**
12. **Timezones & market holidays** — IST sessions, holiday calendar, pre/post-market.
13. **No vendor sandbox** — paper trading must be your own simulator.
14. **Order/product/validity taxonomy** — regular/AMO/CO/iceberg × CNC/MIS/NRML × DAY/IOC/TTL.
15. **Graceful shutdown** — dying mid-order is unsafe.

---

## Part D — Review Checkpoints (the rubric I'll hold each phase to)

When you ask for a review, this is what I'll examine — design accordingly.

**Every review (all phases):**
- Interface boundaries: is the core free of broker/vendor types? Could a second adapter slot in?
- Are interfaces small and consumer-defined? Any "god" interface?
- Money/time handled correctly? Any `float64` money or naive time?
- Errors wrapped and classified (recoverable vs fatal)? Sentinel/typed where branched on?
- Tests present and meaningful on the domain core? Table-driven where appropriate?
- Concurrency: context propagation, no leaked goroutines, race-clean, graceful shutdown?
- No globals; dependencies injected; config not hard-coded; no secrets in code.
- `golangci-lint`/`go vet` clean; idiomatic naming and package boundaries.

**Phase-specific gates:**
- **P1–P2:** Could you delete the SimulatedBroker and drop in a real one with zero core changes?
  Is the event flow clean (no shortcuts bypassing Signal→Order→Fill)?
- **P2:** No lookahead bias; costs modeled; deterministic/reproducible.
- **P3:** New strategy = one interface impl + registration, nothing else touched.
- **P4:** Kite types confined to the adapter; rate limiting present; reconnection handled.
- **P6:** Explicit order state machine; idempotency; partial fills; reconciliation; crash recovery.
- **P7:** Risk is unbypassable and pre-trade; kill switch works.
- **P8:** Observability not retrofitted; CI green with `-race`; secrets externalized; runbook exists.

---

## Part E — Kite Connect Quick Reference (verified, June 2026)

- **Official Go SDK:** `github.com/zerodha/gokiteconnect/v4` — REST client + `ticker` (WebSocket,
  auto-reconnect) + historical candles. Docs: kite.trade/docs/connect/v3.
- **Auth:** API key/secret → login redirect → `request_token` → exchange for `access_token`;
  **access token expires daily** (must re-login each trading day).
- **Rate limits:** orders **10/sec** and **400/min**; **quote 1/sec**; **historical 3/sec**.
- **No sandbox** environment — build your own simulator for paper trading.
- **Order updates:** **postbacks/webhooks** (POST JSON on COMPLETE/CANCEL/REJECTED/UPDATE; UPDATE
  also fires on partial fills) — use as primary update channel, poll as fallback.
- **Static IP required** to place orders via the API (since 1 Apr 2025).
- **Order varieties:** regular, AMO, cover, iceberg. Products: CNC/MIS/NRML. Validity: DAY/IOC/TTL.

---

## Part F — Concepts to Study (so the unknowns become known)

- **Event-driven backtesting** — QuantStart's "event-driven backtester" series (the canonical mental
  model for the Part-0 architecture).
- **Hexagonal / ports-and-adapters architecture** — Alistair Cockburn's original; any Go write-up
  applying it.
- **Go concurrency patterns** — "Go Concurrency Patterns" (Pike), `context` package docs, pipeline &
  fan-in/fan-out patterns, graceful shutdown idioms.
- **Effective Go** + **Go Code Review Comments** — idioms and naming.
- **`log/slog`** (structured logging), **golangci-lint**, **Prometheus client_golang**,
  **OpenTelemetry-Go** basics.
- **Decimal/money handling** — `shopspring/decimal`, or the "integer minor units" approach.
- **Performance metrics for strategies** — Sharpe ratio, max drawdown, CAGR, exposure.
- **Order lifecycle / OMS concepts & FIX-style state machines** (conceptually) — to model order
  states properly.
- **Kite Connect v3 docs** — orders, postbacks, exceptions, WebSocket streaming sections.

---

## Part G — Resume Framing (what this demonstrates)

When done, this is a credible "I built a production-shaped system" story:
- *Designed a broker-agnostic, event-driven trading engine in Go using hexagonal architecture; the
  same strategy core runs identically across backtest, paper, and live execution.*
- *Integrated Zerodha Kite (REST + WebSocket) behind clean interfaces with rate limiting, daily-token
  auth, order-state management, idempotency, and webhook-based reconciliation.*
- *Built a deterministic backtesting engine (commission/slippage modeling, Sharpe/drawdown metrics)
  that doubles as the strategy test harness.*
- *Productionized with structured logging, Prometheus metrics, Dockerized deploy, and CI running the
  race detector + linters.*

Concrete talking points interviewers like: the event-driven seam that enables mode-swapping,
idempotent order handling under network failure, reconciliation, and the concurrency/lifecycle model.

---

## Suggested Cadence

Request a review at the **end of each phase** (and after Phase 1's interfaces *before* implementing
Phase 2 — getting the ports right early is the highest-leverage review). Bring: what you built, the
interfaces involved, and any design decision you were unsure about.
