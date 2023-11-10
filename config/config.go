package config

import (
	"os"
	"time"
)

var IsDev bool = os.Getenv("isDev") == "true"
var IsDebugMode bool = os.Getenv("isDebug") == "true"

const (
	BOOKING_TIMEOUT = time.Second * 30
)
