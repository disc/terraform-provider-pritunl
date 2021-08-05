package provider

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type transport struct {
	underlyingTransport http.RoundTripper
	apiToken            string
	apiSecret           string
	baseUrl             string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	u, err := url.Parse(t.baseUrl + req.URL.String())
	if err != nil {
		return nil, err
	}

	fmt.Println(u)

	req.URL = u

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	timestampNano := strconv.FormatInt(time.Now().UnixNano(), 10)

	nonceMac := hmac.New(md5.New, []byte(t.apiSecret))
	nonceMac.Write([]byte(strings.Join([]string{timestampNano, t.apiToken}, "")))
	nonce := fmt.Sprintf("%x", nonceMac.Sum(nil))
	authString := strings.Join([]string{t.apiToken, timestamp, nonce, strings.ToUpper(req.Method), req.URL.Path}, "&")

	mac := hmac.New(sha256.New, []byte(t.apiSecret))
	mac.Write([]byte(authString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req.Header.Add("Auth-Token", t.apiToken)
	req.Header.Add("Auth-Timestamp", timestamp)
	req.Header.Add("Auth-Nonce", nonce)
	req.Header.Add("Auth-Signature", signature)

	req.Header.Add("Content-Type", "application/json")

	return t.underlyingTransport.RoundTrip(req)
}
