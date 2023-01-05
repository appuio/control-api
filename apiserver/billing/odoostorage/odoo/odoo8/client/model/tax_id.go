package model

import (
	"encoding/json"
	"fmt"
	"math"
)

const (
	firstParameterDefault  = 6
	secondParameterDefault = false
)

// InvoiceLineTaxID represents a VAT entry in an invoice line.
type InvoiceLineTaxID struct {
	// firstParameter is an unknown parameter observed to equal 6
	firstParameter int
	// secondParameter is an unknown parameter observed to equal false
	secondParameter bool

	// ID represents the id of the VAT in an invoice
	ID int
}

// UnmarshalJSON handles deserialization of InvoiceLineTaxID.
func (a *InvoiceLineTaxID) UnmarshalJSON(b []byte) error {
	var params []interface{}
	if err := json.Unmarshal(b, &params); err != nil {
		return err
	}

	isFullNumber := func(n float64) bool {
		_, frac := math.Modf(n)
		return frac == 0
	}

	if len(params) != 3 {
		return fmt.Errorf("expected %d elements in slice, got %d", 3, len(params))
	}
	p1, ok := params[0].(float64)
	if !ok {
		return fmt.Errorf("expected first parameter to be of type float64 (number), got %v", params[0])
	}
	if !isFullNumber(p1) {
		return fmt.Errorf("expected first parameter to be a full number, got %E", p1)
	}
	p2, ok := params[1].(bool)
	if !ok {
		return fmt.Errorf("expected second parameter to be of type bool, got %v", params[1])
	}
	p3, ok := params[2].([]interface{})
	if !ok {
		return fmt.Errorf("expected third parameter to be of type []interface{}, got %v", params[2])
	}
	if len(p3) != 1 {
		return fmt.Errorf("expected third parameter to contain %d element, got %d", 1, len(p3))
	}
	id, ok := p3[0].(float64)
	if !ok {
		return fmt.Errorf("expected the slice of the third parameter to contain a float64 (number), got %v", p3[0])
	}
	if !isFullNumber(id) {
		return fmt.Errorf("expected the slice of the third parameter to contain a full number, got %E", p1)
	}

	a.firstParameter = int(p1)
	a.secondParameter = p2
	a.ID = int(id)
	return nil
}

// MarshalJSON handles serialization of InvoiceLineTaxID.
func (a InvoiceLineTaxID) MarshalJSON() ([]byte, error) {
	p1 := a.firstParameter
	if p1 == 0 {
		p1 = firstParameterDefault
	}
	p2 := a.secondParameter || secondParameterDefault

	return json.Marshal([...]interface{}{p1, p2, [...]int{a.ID}})
}
