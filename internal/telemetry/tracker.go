package telemetry

import (
	"fmt"
)

// Tracker is a telemetry event tracker
type Tracker interface {
	track(event event)
}

type noopTracker struct{}

func (tracker *noopTracker) track(event event) {}

type stdoutTracker struct{}

func (tracker *stdoutTracker) track(event event) {
	fmt.Printf("tracking: %v\n", event)
}

type segmentTracker struct{}

func (tracker *segmentTracker) track(event event) {
	fmt.Printf("tracking: %v\n", event)
}
