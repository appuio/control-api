package client

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession_CreateGenericModel(t *testing.T) {
	numRequests := 0
	uuidGenerator = func() string {
		return "fakeID"
	}
	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		assert.Equal(t, "/web/dataset/call_kw/create", r.RequestURI)

		buf, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, `{
			"id":"fakeID",
			"jsonrpc":"2.0",
			"method":"call",
			"params":{
				"model":"model",
				"method":"create",
				"args":[
					"data"
				],
				"kwargs":{}
			}}`, string(buf))

		w.Header().Set("content-type", "application/json")
		_, err = w.Write([]byte(`{
			"jsonrpc": "2.0",
			"id": "fakeID",
			"result": 221
		}`))
		require.NoError(t, err)
	}))
	defer odooMock.Close()

	// Do request
	u, err := url.Parse(odooMock.URL)
	require.NoError(t, err)
	session := Session{client: &Client{http: http.DefaultClient, parsedURL: u}}
	session.client.http.Transport = newDebugTransport()
	result, err := session.CreateGenericModel(newTestContext(t), "model", "data")
	require.NoError(t, err)
	assert.Equal(t, 221, result)
	assert.Equal(t, 1, numRequests)
}

func TestSession_UpdateGenericModel(t *testing.T) {
	numRequests := 0
	uuidGenerator = func() string {
		return "fakeID"
	}
	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		assert.Equal(t, "/web/dataset/call_kw/write", r.RequestURI)

		buf, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, `{
			"id":"fakeID",
			"jsonrpc":"2.0",
			"method":"call",
			"params":{
				"model":"model",
				"method":"write",
				"args":[
					[1],
					"data"
				],
				"kwargs":{}
			}}`, string(buf))

		w.Header().Set("content-type", "application/json")
		_, err = w.Write([]byte(`{
			"jsonrpc": "2.0",
			"id": "fakeID",
			"result": true
		}`))
		require.NoError(t, err)
	}))
	defer odooMock.Close()

	// Do request
	u, err := url.Parse(odooMock.URL)
	require.NoError(t, err)
	session := Session{client: &Client{http: http.DefaultClient, parsedURL: u}}
	session.client.http.Transport = newDebugTransport()
	err = session.UpdateGenericModel(newTestContext(t), "model", 1, "data")
	require.NoError(t, err)
	assert.Equal(t, 1, numRequests)
}

func TestSession_DeleteGenericModel(t *testing.T) {
	numRequests := 0
	uuidGenerator = func() string {
		return "fakeID"
	}
	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		assert.Equal(t, "/web/dataset/call_kw/unlink", r.RequestURI)

		buf, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, `{
			"id":"fakeID",
			"jsonrpc":"2.0",
			"method":"call",
			"params":{
				"model":"model",
				"method":"unlink",
				"args":[[100]],
				"kwargs":{}
			}}`, string(buf))

		w.Header().Set("content-type", "application/json")
		_, err = w.Write([]byte(`{
			"jsonrpc": "2.0",
			"id": "fakeID",
			"result": true
		}`))
		require.NoError(t, err)
	}))
	defer odooMock.Close()

	// Do request
	u, err := url.Parse(odooMock.URL)
	require.NoError(t, err)
	session := Session{client: &Client{http: http.DefaultClient, parsedURL: u}}
	session.client.http.Transport = newDebugTransport()
	err = session.DeleteGenericModel(newTestContext(t), "model", []int{100})
	require.NoError(t, err)
	assert.Equal(t, 1, numRequests)
}
