package config

import (
	"os"
	"time"
)

var IsDev bool = os.Getenv("isDev") == "true"

const (
	BOOKING_TIMEOUT = time.Second * 30
)
