package types

type Money struct {
	Amount   int64
	Currency Currency
}
type Currency string

const (
	INR Currency = "INR"
)
