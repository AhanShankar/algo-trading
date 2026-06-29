package trading

import (
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
)

type Order struct {
	ID          string
	Instrument  Instrument
	Side        Side
	Quantity    int64
	Price       Money
	Type        OrderType
	ProductType ProductType
	Validity    OrderValidity
	State       OrderState
}

type OrderType string

const (
	Market   OrderType = "MARKET"
	Limit    OrderType = "LIMIT"
	StopLoss OrderType = "STOPLOSS"
)

type ProductType string

const (
	Delivery ProductType = "DELIVERY"
	Intraday ProductType = "INTRADAY"
	Normal   ProductType = "NORMAL"
)

type OrderValidity string

const (
	Day       OrderValidity = "DAY"
	Immediate OrderValidity = "IMMEDIATE"
)

type OrderState string

const (
	Pending         OrderState = "PENDING"
	Open            OrderState = "OPEN"
	PartiallyFilled OrderState = "PARTIAL"
	Complete        OrderState = "COMPLETE"
	Rejected        OrderState = "REJECTED"
	Cancelled       OrderState = "CANCELLED"
)

func (o Order) GenerateID() string {
	timestamp := time.Now()
	id := strings.Join([]string{
		timestamp.Format(time.RFC3339),
		string(o.Side),
		o.Instrument.Ticker,
		strconv.FormatInt(o.Quantity, 10),
		randomSuffix(),
	}, "-")
	o.ID = id
	return o.ID
}

// randomSuffix returns a 5-character A–Z string used to disambiguate IDs
// generated within the same second. math/rand/v2's top-level functions are
// auto-seeded and safe for concurrent use, so there is no Seed call to get
// wrong (the previous per-call time-based seeding could collide).
func randomSuffix() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 5)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}
	return string(b)
}
