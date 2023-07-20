package helper

import "os"

var (
	ConsumerCount   = 0
	ExitAMQP        = make(chan os.Signal, 1)
	ExitConsumer    = make(chan bool)
	ExitHTTP        = make(chan bool)
	ExitConcurrency = make(chan bool)
)
