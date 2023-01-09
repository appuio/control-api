package client

// SearchReadModel is used as "params" in requests to "dataset/search_read" endpoints.
type SearchReadModel struct {
	Model  string   `json:"model,omitempty"`
	Domain []Filter `json:"domain,omitempty"`
	Fields []string `json:"fields,omitempty"`
	Limit  int      `json:"limit,omitempty"`
	Offset int      `json:"offset,omitempty"`
}

// Filter to use in queries, usually in the format of
// [predicate, operator, value], eg ["employee_id.user_id.id", "=", 123]
type Filter any

// Method identifies the type of write operation.
type Method string

const (
	// MethodWrite is used to update existing records.
	MethodWrite Method = "write"
	// MethodCreate is used to create new records.
	MethodCreate Method = "create"
	// MethodRead is used to read records.
	MethodRead Method = "read"
	// MethodDelete is used to delete existing records.
	MethodDelete Method = "unlink"
)

// WriteModel is used as "params" in requests to "dataset/create", "dataset/write" or "dataset/unlinke" endpoints.
type WriteModel struct {
	Model  string `json:"model"`
	Method Method `json:"method"`
	// Args contains the record to create or update.
	// If Method is MethodCreate, then the slice may contain a single entity without an ID parameter.
	// Example:
	//  Args[0] = {Name: "New Name"}
	// If Method is MethodWrite, then the first item has to be an array of the numeric ID of the existing record.
	// Example:
	//  Args[0] = [221]
	//  Args[1] = {Name: "Updated Name"}
	Args []any `json:"args"`
	// KWArgs is an additional object required to be non-nil, otherwise the request simply fails.
	// In most cases it's enough to set it to `map[string]any{}`.
	KWArgs map[string]any `json:"kwargs"`
}
