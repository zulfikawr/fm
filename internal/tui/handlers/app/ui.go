package app

import (
	"time"
)

// TickMsg is sent on each tick for animations
type TickMsg struct {
	Time time.Time
}
