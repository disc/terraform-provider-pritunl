package pritunl

import (
	"net/http"
	"net/url"
	"path"
)

type transport struct {
	underlyingTransport http.RoundTripper
	baseUrl             string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == "" {
		u, err := url.Parse(t.baseUrl)
		if err != nil {
			return nil, err
		}

		u.Path = path.Join(u.Path, req.URL.Path)
		req.URL = u
	}

	req.Header.Add("Content-Type", "application/json")
	return t.underlyingTransport.RoundTrip(req)
}
