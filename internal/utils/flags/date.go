package flags

import (
	"fmt"
	"time"
)

const (
	dateFormatTZ = "2006-01-02T15:04:05.000-0700"
	dateFormat   = "2006-01-02T15:04:05.000"

	dateFormatSecondsTZ = "2006-01-02T15:04:05-0700"
	dateFormatSeconds   = "2006-01-02T15:04:05"
	dateFormatMinutesTZ = "2006-01-02T15:04-0700"
	dateFormatMinutes   = "2006-01-02T15:04"
	dateFormatHoursTZ   = "2006-01-02T15-0700"
	dateFormatHours     = "2006-01-02T15"
	dateFormatDaysTZ    = "2006-01-02-0700"
	dateFormatDays      = "2006-01-02"
)

// Date is a date flag
type Date struct {
	Time time.Time
}

// Type returns the date flag type
func (d Date) Type() string {
	return "Date"
}

func (d Date) String() string {
	if d.Time.IsZero() {
		return ""
	}
	return d.Time.Format(dateFormatTZ)
}

// Set adds new values to the EnumSetValue
func (d *Date) Set(val string) error {
	t, err := parseTime(val)
	if err != nil {
		return err
	}

	d.Time = t
	return nil
}

func parseTime(val string) (time.Time, error) {
	if t, err := time.Parse(dateFormatTZ, val); err == nil {
		return t, nil
	}
	if t, err := time.Parse(dateFormat, val); err == nil {
		return t, nil
	}
	if t, err := time.Parse(dateFormatSecondsTZ, val); err == nil {
		return t, nil
	}
	if t, err := time.Parse(dateFormatSeconds, val); err == nil {
		return t, nil
	}
	if t, err := time.Parse(dateFormatMinutesTZ, val); err == nil {
		return t, nil
	}
	if t, err := time.Parse(dateFormatMinutes, val); err == nil {
		return t, nil
	}
	if t, err := time.Parse(dateFormatHoursTZ, val); err == nil {
		return t, nil
	}
	if t, err := time.Parse(dateFormatHours, val); err == nil {
		return t, nil
	}
	if t, err := time.Parse(dateFormatDaysTZ, val); err == nil {
		return t, nil
	}
	if t, err := time.Parse(dateFormatDays, val); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unrecognized date string: %s", val)
}
