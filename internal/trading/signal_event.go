package trading

import "time"

type SignalEvent struct {
	Instrument Instrument
	Direction  SignalDirection
	Timestamp  time.Time
}
type SignalDirection string

const (
	Long  SignalDirection = "LONG"
	Short SignalDirection = "SHORT"
	Exit  SignalDirection = "EXIT"
)
