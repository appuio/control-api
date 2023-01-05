package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"
)

// JSONRPCRequest represents a generic json-rpc request
type JSONRPCRequest struct {
	// ID should be a randomly generated value, either as a string or int.
	// The server will return this value in the response.
	ID string `json:"id,omitempty"`

	// JSONRPC is always set to "2.0"
	JSONRPC string `json:"jsonrpc,omitempty"`

	// Method to call, usually just "call"
	Method string `json:"method,omitempty"`

	// Params includes the actual request payload.
	Params interface{} `json:"params,omitempty"`
}

var uuidGenerator = uuid.NewString

// NewJSONRPCRequest returns a JSON RPC request with its protocol fields populated:
//
// * "id" will be set to a random UUID
// * "jsonrpc" will be set to "2.0"
// * "method" will be set to "call"
// * "params" will be set to whatever was passed in
func NewJSONRPCRequest(params interface{}) *JSONRPCRequest {
	return &JSONRPCRequest{
		ID:      uuidGenerator(),
		JSONRPC: "2.0",
		Method:  "call",
		Params:  params,
	}
}

// Encode encodes the request as JSON in a buffer and returns the buffer.
func (r *JSONRPCRequest) Encode() (io.Reader, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(r); err != nil {
		return nil, err
	}

	return buf, nil
}

// JSONRPCResponse holds the JSONRPC response.
type JSONRPCResponse struct {
	// ID that was sent with the request
	ID string `json:"id,omitempty"`
	// JSONRPC is always set to "2.0"
	JSONRPC string `json:"jsonrpc,omitempty"`
	// Result payload
	Result *json.RawMessage `json:"result,omitempty"`

	// Optional error field
	Error *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError holds error information.
type JSONRPCError struct {
	Message string                 `json:"message,omitempty"`
	Code    int                    `json:"code,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// DecodeResult takes a buffer, decodes the intermediate JSONRPCResponse and then the contained "result" field into "result".
func DecodeResult(buf io.Reader, result interface{}) error {
	// Decode intermediate
	var res JSONRPCResponse
	if err := json.NewDecoder(buf).Decode(&res); err != nil {
		return fmt.Errorf("decode intermediate: %w", err)
	}
	if res.Error != nil {
		return fmt.Errorf("%s: %s", res.Error.Message, res.Error.Data["message"])
	}

	return json.Unmarshal(*res.Result, result)
}

func newEncodingRequestError(err error) error {
	return fmt.Errorf("encoding request: %w", err)
}

func newCreatingRequestError(err error) error {
	return fmt.Errorf("creating request: %w", err)
}
