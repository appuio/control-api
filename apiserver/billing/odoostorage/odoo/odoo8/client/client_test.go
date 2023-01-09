package client

import (
	"context"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/go-logr/zapr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"k8s.io/klog/v2"
)

func TestClient_parseURL(t *testing.T) {
	tests := map[string]struct {
		givenURL         string
		expectedPassword string
		expectedUsername string
		expectedDBName   string
		expectedError    string
	}{
		"GivenURLWithoutUserInfo_ThenExpectError": {
			givenURL:      "https://host:80/db",
			expectedError: "missing username and password in URL",
		},
		"GivenURLWithoutPassword_ThenExpectError": {
			givenURL:      "https://user@host:80/db",
			expectedError: "missing password in URL",
		},
		"GivenURLWithoutDB_ThenExpectError": {
			givenURL:      "https://user:pass@host:80/",
			expectedError: "missing db name in URL path",
		},
		"GivenValidURL_ThenExpectParsedProperties": {
			givenURL:         "https://user:pass@host:80/db-name",
			expectedUsername: "user",
			expectedPassword: "pass",
			expectedDBName:   "db-name",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Client{}
			err := c.parseOdooURL(tc.givenURL)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedDBName, c.db)
			assert.Equal(t, tc.expectedUsername, c.username)
			assert.Equal(t, tc.expectedPassword, c.password)
		})
	}
}

func TestClient_Login_Success(t *testing.T) {
	var (
		numRequests  = 0
		testLogin    = uuid.NewString()
		testPassword = uuid.NewString()
		testUID      = rand.Int()
		testSID      = uuid.NewString()
	)

	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		assert.Equal(t, "/web/session/authenticate", r.RequestURI)

		b, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)
		body := string(b)
		assert.Contains(t, body, `"db":"TestDB"`)
		assert.Contains(t, body, `"login":"`+testLogin+`"`)
		assert.Contains(t, body, `"password":"`+testPassword+`"`)

		w.Header().Set("content-type", "application/json")
		_, err = w.Write([]byte(`{
			"id": "1337",
			"jsonrpc": "2.0",
			"result": {
				"company_id": 1,
				"db": "TestDB",
				"session_id": "` + testSID + `",
				"uid": ` + strconv.Itoa(testUID) + `,
				"user_context": {
					"lang": "en_US",
					"tz": "Europe/Zurich",
					"uid": ` + strconv.Itoa(testUID) + `
				},
				"username": "` + testLogin + `"
			}
		}`))
		require.NoError(t, err)
	}))
	defer odooMock.Close()

	// login
	u := newTestURL(t, odooMock.URL, testLogin, testPassword, "TestDB")
	session, err := Open(newTestContext(t), u, ClientOptions{UseDebugLogger: true})
	require.NoError(t, err)
	assert.Equal(t, testUID, session.UID)
	assert.Equal(t, testSID, session.SessionID)
	assert.Equal(t, 1, numRequests)
}

func TestLogin_BadCredentials(t *testing.T) {
	var (
		numRequests  = 0
		testLogin    = uuid.NewString()
		testPassword = uuid.NewString()
		testSID      = uuid.NewString()
	)

	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		assert.Equal(t, "/web/session/authenticate", r.RequestURI)
		w.Header().Set("content-type", "application/json")
		_, err := w.Write([]byte(`{
			"id": "1337",
			"jsonrpc": "2.0",
			"result": {
				"company_id": null,
				"db": "TestDB",
				"session_id": "` + testSID + `",
				"uid": false,
				"user_context": {},
				"username": "` + testLogin + `"
			}
		}`))
		require.NoError(t, err)
	}))
	defer odooMock.Close()

	// Do request
	u := newTestURL(t, odooMock.URL, testLogin, testPassword, "TestDB")
	session, err := Open(newTestContext(t), u, ClientOptions{UseDebugLogger: true})
	require.EqualError(t, err, "invalid credentials")
	assert.Nil(t, session)
	assert.Equal(t, 1, numRequests)
}

func TestLogin_BadResponse(t *testing.T) {
	numRequests := 0
	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++

		w.Header().Set("content-type", "application/json")
		_, err := w.Write([]byte(`{
			"jsonrpc": "2.0",
			"id": "xxx",
			"error": {
			  "message": "Odoo Server Error",
			  "code": 200,
			  "data": {
				"debug": "Traceback xxx",
				"message": "",
				"name": "werkzeug.exceptions.Foo",
				"arguments": []
			  }
			}
		  }`))
		require.NoError(t, err)
	}))
	defer odooMock.Close()

	// Do request
	u := newTestURL(t, odooMock.URL, "irrelevant", "irrelevant", "TestDB")
	session, err := Open(newTestContext(t), u, ClientOptions{UseDebugLogger: true})
	require.EqualError(t, err, "error from Odoo: &{Odoo Server Error 200 map[arguments:[] debug:Traceback xxx message: name:werkzeug.exceptions.Foo]}")
	assert.Nil(t, session)
	assert.Equal(t, 1, numRequests)
}

func newTestContext(t *testing.T) context.Context {
	zlogger := zaptest.NewLogger(t, zaptest.Level(zapcore.Level(-2)))
	return klog.NewContext(context.Background(), zapr.NewLogger(zlogger))
}

func newTestURL(t *testing.T, baseURL, username, password, db string) string {
	parsed, err := url.Parse(baseURL)
	require.NoError(t, err)
	user := url.UserPassword(username, password)
	parsed.User = user
	parsed.Path = db
	return parsed.String()
}
