# algo-trading

A broker-agnostic algorithmic trading system in Go. Built to learn — not to get rich.

## What this is

An event-driven trading engine that runs the same strategy code across backtesting, paper trading, and live trading. The broker (currently Zerodha/Kite) is hidden behind interfaces, so swapping it out touches zero strategy or engine code.

## Architecture

The core idea is a pipeline of events:

```
Market Data → [MarketEvent] → Strategy → [SignalEvent] → Risk → [OrderEvent] → Broker → [FillEvent] → Portfolio
```

Backtest, paper, and live modes differ only in where market events come from and what executes orders. Everything else is shared.

This is hexagonal architecture (ports & adapters) — the domain core depends only on interfaces; Kite, CSV readers, and simulated brokers are swappable adapters.

## Build plan

| Phase | What |
|-------|------|
| 0 | Core domain types (Money, Instrument, Tick, Candle, Order) ✅ |
| 1 | Core interfaces (ports) + event engine |
| 2 | Backtesting engine (CSV data + simulated broker) |
| 3 | Strategy framework + indicators (SMA, EMA, RSI) |
| 4 | Kite market data adapter (REST + WebSocket) |
| 5 | Paper trading (live data + simulated execution) |
| 6 | Live trading (real Kite broker adapter + OMS) |
| 7 | Risk management + kill switch |
| 8 | Production hardening (observability, Docker, CI/CD) |

## Stack

- **Language:** Go 1.25
- **Broker:** Zerodha Kite Connect (`gokiteconnect/v4`)
- **Observability:** `log/slog`, Prometheus (planned)
- **CI:** GitHub Actions (planned)

## Running

```bash
go test ./...
```

> Full run instructions coming in Phase 2 once the backtester is wired up.
