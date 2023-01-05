package client

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/go-logr/logr"
)

type debugTransport struct {
	pwRe *regexp.Regexp
}

func newDebugTransport() *debugTransport {
	return &debugTransport{
		pwRe: regexp.MustCompile(`("password":\s?").+("[,}])`),
	}
}

func (c Client) useDebugLogger(enabled bool) {
	if enabled {
		c.http.Transport = newDebugTransport()
	}
}

// RoundTrip implements http.RoundTripper.
func (t *debugTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	logger := logr.FromContextOrDiscard(r.Context())
	if r.Body != nil {
		reqBody, _ := r.GetBody()
		defer reqBody.Close()
		buf, _ := ioutil.ReadAll(reqBody)
		buf = t.pwRe.ReplaceAll(buf, []byte(`$1[confidential]$2`))
		logger.V(2).Info(fmt.Sprintf("%s %s ---> %s", r.Method, r.URL.Path, string(buf)))
	}

	res, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
		buf, _ := ioutil.ReadAll(res.Body)
		logger.V(2).Info(fmt.Sprintf("%s %s <--- %s", r.Method, r.URL.Path, string(buf)))
		res.Body = io.NopCloser(bytes.NewReader(buf))
	}

	return res, err
}
