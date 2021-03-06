package config

import (
	"time"
)

const (
	StateTransmissionInterval           = 200 * time.Millisecond
	ElevatorTimeout                     = 10000 * time.Millisecond
	LastUpdateInterval                  = 2000 * time.Millisecond
	ActiveElevatorsTransmissionInterval = 500 * time.Millisecond
	TransmissionPort                    = 19569
	PollRate                            = 20 * time.Millisecond
)
