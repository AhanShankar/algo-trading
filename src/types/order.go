package types

import (
	"strconv"
	"time"

	"github.com/AhanShankar/algo-trading/src/utils"
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

func (o Order) GenerateId() string {
	timestamp := time.Now()
	return timestamp.Format(time.RFC3339) + string(o.Side) + o.Instrument.Ticker + strconv.FormatInt(o.Quantity, 10) + utils.RandomCapsString()
}
