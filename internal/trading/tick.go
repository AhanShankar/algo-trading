package trading

import "time"

type Tick struct {
	Instrument    Instrument
	Price         Money
	Quantity      int64
	Volume        int64
	BestBuyPrice  Money
	BestSellPrice Money
	Timestamp     time.Time
	DayOpen       Money
	DayClose      Money
	DayHigh       Money
	DayLow        Money
}
