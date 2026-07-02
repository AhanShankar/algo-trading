package engine

import (
	"context"

	"github.com/AhanShankar/algo-trading/internal/trading"
)

type Broker interface {
	// PlaceOrder submits a batch of orders and returns one result per order,
	// correlated by ClientID. Fills arrive asynchronously and are not part of
	// this return.
	PlaceOrder(ctx context.Context, orders []trading.Order) []trading.OrderResult
	GetOrderStatus(ctx context.Context, order trading.Order) (trading.Order, error)
}
