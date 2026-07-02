package engine

import "github.com/AhanShankar/algo-trading/internal/trading"

type Store interface {
	PersistOrders(orders []trading.Order) error
	PersistFills(fills []trading.Fill) error
	PersistCandles(candles []trading.Candle) error
}
