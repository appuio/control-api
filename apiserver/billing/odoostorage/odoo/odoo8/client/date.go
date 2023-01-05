package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

const (
	// DateFormat only yields year, month, day.
	DateFormat = "2006-01-02"
	// TimeFormat only yields hour, minute, seconds in 24-h format.
	TimeFormat = "15:04:05"
	// DateTimeFormat combines DateFormat with TimeFormat separated by space.
	DateTimeFormat = DateFormat + " " + TimeFormat
)

// Date is an Odoo-specific format of a timestamp
type Date time.Time

// String formats date using DateTimeFormat.
func (d *Date) String() string {
	t := time.Time(*d)
	return t.Format(DateTimeFormat)
}

// MarshalJSON implements json.Marshaler.
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Date) UnmarshalJSON(b []byte) error {
	ts := bytes.Trim(b, `"`)
	var f bool
	if err := json.Unmarshal(b, &f); err == nil || string(b) == "false" {
		return nil
	}
	// try parsing date + time
	t, dateTimeErr := time.Parse(DateTimeFormat, string(ts))
	if dateTimeErr != nil {
		// second attempt parsing date only
		t, dateTimeErr = time.Parse(DateFormat, string(ts))
		if dateTimeErr != nil {
			return dateTimeErr
		}
	}

	*d = Date(t)
	return nil
}

// IsZero returns true if Date is nil or Time.IsZero()
func (d *Date) IsZero() bool {
	return d == nil || d.ToTime().IsZero()
}

// ToTime returns a time.Time representation.
func (d Date) ToTime() time.Time {
	return time.Time(d)
}
