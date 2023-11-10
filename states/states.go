package states

type State int

const (
	INIT State = iota
	PASSENGER
	FARE
	PICKUP
	DROPOFF
	PENDING_DRIVER
	PENDING_PICKUP
	DRIVER_STATE

	// shorten it to fit telegram callback data limit
	ACCEPT_BOOKING     = "A"
	CANCEL_PICKUP      = "B"
	SEND_ARRIVAL       = "C"
	PAX_CANCEL_BOOKING = "D"
	RE_CREATE          = "E"
	PAX_COMPLETED      = "F"
)
