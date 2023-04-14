package model

import "encoding/json"

type CategoryIDs []int

func (t CategoryIDs) MarshalJSON() ([]byte, error) {
	// Values observed on test and prod instances.
	// Seems to be some kinda update join table command.
	// Also observed in other n..m relations.
	return json.Marshal([][]any{{6, false, []int(t)}})
}
