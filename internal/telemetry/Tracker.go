package telemetry

import (
	"fmt"
)

type Tracker interface {
	Track(event Event)
}

type TrackerType string

const (
	//TrackerTypeViperKey used to get the trackertype from viper
	TrackerTypeViperKey string = "trackerType"
	//User deliberately selects this option
	OnSelected TrackerType = ""
	//User does not select an option
	OnDefault TrackerType = "OnDefault"
	//User selects stdout
	STDOut TrackerType = "stdout"
	//User disables tracking
	Off TrackerType = "off"
)

var (
	trackerType TrackerType
)

// NewTracker constructs a new tracking service
func NewTracker(profileString string, configString string) (Tracker, error) {
	trackerType = TrackerType(profileString)
	if TrackerType(configString) != OnDefault {
		trackerType = TrackerType(configString)
	}
	switch trackerType {
	case OnSelected:
		return &segmentTracker{}, nil
	case OnDefault:
		return &segmentTracker{}, nil
	case STDOut:
		return &stdoutTracker{}, nil
	case Off:
		return &noopTracker{}, nil
	}
	return nil, fmt.Errorf("%q is not a recognized tracking config type", trackerType)
}

// GetTrackerConfig gets the current tracker config.
func GetTrackerConfig() TrackerType {
	return trackerType
}

type noopTracker struct{}

func (tracker *noopTracker) Track(event Event) {}

type stdoutTracker struct{}

func (tracker *stdoutTracker) Track(event Event) {
	fmt.Printf("tracking: %v", event)
}

type segmentTracker struct{}

func (tracker *segmentTracker) Track(event Event) {
	fmt.Printf("tracking: %v", event)
}
