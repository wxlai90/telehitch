package config

import "os"

var IsDev bool = os.Getenv("isDev") == "true"
