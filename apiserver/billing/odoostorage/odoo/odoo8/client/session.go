package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	// ErrInvalidCredentials is an error that indicates an authentication error due to missing or invalid credentials.
	ErrInvalidCredentials = errors.New("invalid credentials")
)

//go:generate go run go.uber.org/mock/mockgen -destination=./clientmock/$GOFILE -package clientmock . QueryExecutor

// QueryExecutor runs queries against Odoo API.
type QueryExecutor interface {
	// SearchGenericModel accepts a SearchReadModel and unmarshal the response into the given pointer.
	// Depending on the JSON fields returned a custom json.Unmarshaler needs to be written since Odoo sets undefined fields to `false` instead of null.
	SearchGenericModel(ctx context.Context, model SearchReadModel, into any) error
	// CreateGenericModel accepts a payload and executes a query to create the new data record.
	CreateGenericModel(ctx context.Context, model string, data any) (int, error)
	// UpdateGenericModel accepts a payload and executes a query to update an existing data record.
	UpdateGenericModel(ctx context.Context, model string, ids []int, data any) error
	// DeleteGenericModel accepts a model identifier and data records IDs as payload and executes a query to delete multiple existing data records.
	// At least one ID is required.
	DeleteGenericModel(ctx context.Context, model string, ids []int) error
	// ExecuteQuery runs a generic JSONRPC query with the given model as payload and deserializes the response.
	ExecuteQuery(ctx context.Context, path string, model any, into any) error
}

// Session information
type Session struct {
	// SessionID is the session SessionID.
	// Is always set, no matter the authentication outcome.
	SessionID string `json:"session_id,omitempty"`
	// UID is the user's ID as an int, or the boolean `false` if authentication failed.
	UID    int `json:"uid,omitempty"`
	client *Client
}

// SearchGenericModel implements QueryExecutor.
func (s *Session) SearchGenericModel(ctx context.Context, model SearchReadModel, into any) error {
	return s.ExecuteQuery(ctx, "/web/dataset/search_read", model, into)
}

// CreateGenericModel implements QueryExecutor.
func (s *Session) CreateGenericModel(ctx context.Context, model string, data any) (int, error) {
	payload := WriteModel{
		Model:  model,
		Method: MethodCreate,
		Args:   []any{data},
		KWArgs: map[string]any{}, // set to non-null when serializing
	}
	resultID := 0
	err := s.ExecuteQuery(ctx, "/web/dataset/call_kw/create", payload, &resultID)
	return resultID, err
}

// UpdateGenericModel implements QueryExecutor.
func (s *Session) UpdateGenericModel(ctx context.Context, model string, ids []int, data any) error {
	if len(ids) == 0 {
		return fmt.Errorf("ids are required")
	}
	for _, id := range ids {
		if id == 0 {
			return fmt.Errorf("ids can't be zero: %v", ids)
		}
	}

	payload := WriteModel{
		Model:  model,
		Method: MethodWrite,
		Args: []any{
			ids,
			data,
		},
		KWArgs: map[string]any{}, // set to non-null when serializing
	}
	updated := false
	err := s.ExecuteQuery(ctx, "/web/dataset/call_kw/write", payload, &updated)
	return err
}

// DeleteGenericModel implements QueryExecutor.
func (s *Session) DeleteGenericModel(ctx context.Context, model string, ids []int) error {
	if len(ids) == 0 {
		return fmt.Errorf("slice of ID(s) is required")
	}
	for i, id := range ids {
		if id == 0 {
			return fmt.Errorf("id cannot be zero (index: %d)", i)
		}
	}
	payload := WriteModel{
		Model:  model,
		Method: MethodDelete,
		Args:   []any{ids},
		KWArgs: map[string]any{}, // set to non-null when serializing
	}
	deleted := false
	err := s.ExecuteQuery(ctx, "/web/dataset/call_kw/unlink", payload, &deleted)
	return err
}

// ExecuteQuery implements QueryExecutor.
func (s *Session) ExecuteQuery(ctx context.Context, path string, model any, into any) error {
	body, err := NewJSONRPCRequest(&model).Encode()
	if err != nil {
		return newEncodingRequestError(err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.client.parsedURL.String()+path, body)
	if err != nil {
		return newCreatingRequestError(err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("cookie", "session_id="+s.SessionID)

	resp, err := s.sendRequest(req)
	if err != nil {
		return err
	}
	return s.unmarshalResponse(resp.Body, into)
}

func (s *Session) sendRequest(req *http.Request) (*http.Response, error) {
	res, err := s.client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending HTTP request: %w", err)
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected HTTP status 200 OK, got %s", res.Status)
	}
	return res, nil
}

func (s *Session) unmarshalResponse(body io.ReadCloser, into any) error {
	defer body.Close()
	if err := DecodeResult(body, into); err != nil {
		return fmt.Errorf("decoding result: %w", err)
	}
	return nil
}
