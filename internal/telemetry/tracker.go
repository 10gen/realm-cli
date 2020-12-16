package telemetry

import (
	"fmt"
	"time"
)

// Tracker is a telemetry event tracker
type Tracker interface {
	Track(event event)
}

type noopTracker struct{}

func (tracker *noopTracker) Track(event event) {}

type stdoutTracker struct{}

func (tracker *stdoutTracker) Track(event event) {
	fmt.Printf(
		"%s UTC TELEM %s: %s%v\n",
		event.time.In(time.UTC).Format("15:04:05"),
		event.command,
		event.eventType,
		event.data,
	)
}

type segmentTracker struct{}

// TODO(REALMC-7243): use Segment sdk to send events through client
func (tracker *segmentTracker) Track(event event) {
	fmt.Printf(
		"%s UTC TELEM %s: %s%v\n",
		event.time.In(time.UTC).Format("15:04:05"),
		event.command,
		event.eventType,
		event.data,
	)
}
