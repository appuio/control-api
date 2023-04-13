package model

import "encoding/json"

type CategoryIDs []int

func (t CategoryIDs) MarshalJSON() ([]byte, error) {
	// Values observed on test and prod instances.
	// I neither know or care what the fuck they mean.
	return json.Marshal([][]any{{6, false, []int(t)}})
}
