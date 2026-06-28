package types

type Instrument struct {
	Ticker   string
	Exchange string
	Type     InstrumentType
}

type InstrumentType string

const (
	Equity InstrumentType = "EQUITY"
)
