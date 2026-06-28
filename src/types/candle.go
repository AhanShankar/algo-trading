package types

import "time"

type Candle struct {
	Open       Money
	Close      Money
	Low        Money
	High       Money
	Volume     int64
	Instrument Instrument
	Timestamp  time.Time
	Interval   time.Duration
}
