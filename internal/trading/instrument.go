package trading

type Instrument struct {
	Ticker   string
	Exchange string
	Type     InstrumentType
}

type InstrumentType string

const (
	Equity InstrumentType = "EQUITY"
)

func (i Instrument) Key() string {
	return i.Exchange + ":" + i.Ticker
}
