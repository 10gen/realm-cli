package telemetry

import (
	"fmt"
)

// Tracker logs events
type Tracker interface {
	Track(event Event)
}

// NewTracker constructs a new tracking service
func NewTracker(mode Mode) Tracker {
	switch mode {
	case OnSelected:
		return &segmentTracker{}
	case OnDefault:
		return &segmentTracker{}
	case STDOut:
		return &stdoutTracker{}
	}
	return &noopTracker{}
}

type noopTracker struct{}

func (tracker *noopTracker) Track(event Event) {}

type stdoutTracker struct{}

func (tracker *stdoutTracker) Track(event Event) {
	fmt.Sprintf("tracking: %v\n", event)
}

type segmentTracker struct{}

func (tracker *segmentTracker) Track(event Event) {
	fmt.Printf("tracking: %v\n", event)
}
