package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
)

// OdooCompositeID represents a composite ID used in the Odoo API.
// It is a tuple of an ID and a name.
// The Valid field is set if Odoo returns a tuple.
// Valid is unset if Odoo returns `false` in the response instead of the expected tuple.
// This most likely means that the field is not set (== null).
type OdooCompositeID struct {
	// Valid is true if the OdooCompositeID is not null.
	// Odoo returns false if a field is not set. Due to how json unmarshalling works,
	// and the lack of pass-by-reference in Go, we can't set the field to nil.
	// https://dave.cheney.net/2017/04/29/there-is-no-pass-by-reference-in-go
	Valid bool
	// ID is the data record identifier.
	ID int
	// Name is a human-readable description.
	Name string
}

// UnmarshalJSON handles deserialization of OdooCompositeID.
func (t *OdooCompositeID) UnmarshalJSON(b []byte) error {
	// Odoo returns false (not null) if a field is not set.
	if bytes.Equal(b, []byte("false")) {
		return nil
	}

	var values []any
	if err := json.Unmarshal(b, &values); err != nil {
		return err
	}

	isFullNumber := func(n float64) bool {
		_, frac := math.Modf(n)
		return frac == 0
	}

	if len(values) != 2 {
		return fmt.Errorf("expected %d elements in slice, got %d", 2, len(values))
	}

	tID, ok := values[0].(float64)
	if !ok {
		return fmt.Errorf("expected first value to be of type float64 (number), got %v", values[0])
	}
	if !isFullNumber(tID) {
		return fmt.Errorf("expected first value to be a full number, got %E", tID)
	}

	tName, ok := values[1].(string)
	if !ok {
		return fmt.Errorf("expected second value to be of type string, got %v", values[1])
	}

	t.ID = int(tID)
	t.Name = tName
	t.Valid = true
	return nil
}

// MarshalJSON handles serialization of OdooCompositeID.
func (t OdooCompositeID) MarshalJSON() ([]byte, error) {
	return json.Marshal([...]any{t.ID, t.Name})
}
