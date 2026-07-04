package trading

import "time"

type Fill struct {
	OrderID      string
	Quantity     int64
	Price        Money
	Timestamp    time.Time
	BrokerageFee Money
	Instrument   Instrument
	Side         Side
}
