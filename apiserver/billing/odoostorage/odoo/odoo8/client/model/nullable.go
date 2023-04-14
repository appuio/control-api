package model

import (
	"bytes"
	"encoding/json"
)

// Nullable represents a nullable value.
// The Valid field is set if Odoo returns a value.
// Valid is unset if Odoo returns `false` or `null` in the response instead of the expected value.
type Nullable[T any] struct {
	Value T
	Valid bool
}

// NewNullable creates a new Nullable[T] with the given value.
func NewNullable[T any](v T) Nullable[T] {
	return Nullable[T]{
		Value: v,
		Valid: true,
	}
}

func (t *Nullable[T]) UnmarshalJSON(b []byte) error {
	// Odoo returns false (not null) if a field is not set.
	if bytes.Equal(b, []byte("false")) {
		return nil
	}
	if bytes.Equal(b, []byte("null")) {
		return nil
	}

	t.Valid = true
	return json.Unmarshal(b, &t.Value)
}

func (t Nullable[T]) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return []byte("false"), nil
	}
	return json.Marshal(t.Value)
}
