package mock

import (
	"github.com/10gen/realm-cli/internal/telemetry"
)

// TelemetryService is a mocked telemetry service
type TelemetryService struct {
	telemetry.Service
	TrackEventFn func(eventType telemetry.EventType, data ...telemetry.EventData)
	CloseFn      func()
}

// TrackEvent calls the mocked TrackEvent implementation if provided,
// otherwise the call falls back to the underlying telemetry.Service implementation.
// NOTE: this may panic if the underlying telemetry.Service is left undefined
func (s TelemetryService) TrackEvent(eventType telemetry.EventType, data ...telemetry.EventData) {
	if s.TrackEventFn != nil {
		s.TrackEventFn(eventType, data...)
		return
	}
	s.Service.TrackEvent(eventType, data...)
}

// Close calls the mocked Close implementation if provided,
// otherwise the call falls back to the underlying telemetry.Service implementation.
// NOTE: this may panic if the underlying telemetry.Service is left undefined
func (s TelemetryService) Close() {
	if s.CloseFn != nil {
		s.CloseFn()
		return
	}
	s.Service.Close()
}
