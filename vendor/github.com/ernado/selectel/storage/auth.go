package storage

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	authURL          = "https://auth.selcdn.ru/"
	authUserHeader   = "X-Auth-User"
	authKeyHeader    = "X-Auth-Key"
	authExpireHeader = "X-Expire-Auth-Token"
	storageURLHeader = "X-Storage-Url"
	// tokenDurationAdd used to reduce duration of token
	// to re-auth before token gets expired
	tokenDurationAdd = 10 * time.Second
)

var (
	// ErrorAuth occurs when client is unable to authenticate
	ErrorAuth = errors.New("Authentication error")
	// ErrorBadCredentials occurs when incorrect user/key provided
	ErrorBadCredentials = errors.New("Bad auth credentials provided")
)

// Token returns current auth token
func (c *Client) Token() string {
	return c.token
}

// Auth performs authentication to selectel and stores token and storage url
func (c *Client) Auth(user, key string) error {
	if blank(user) || blank(key) {
		return ErrorBadCredentials
	}

	request, _ := http.NewRequest(getMethod, authURL, nil)
	request.Header.Add(authUserHeader, user)
	request.Header.Add(authKeyHeader, key)

	res, err := c.do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return ErrorAuth
	}
	expire, err := strconv.Atoi(res.Header.Get(authExpireHeader))
	if err != nil {
		return err
	}

	c.tokenExpire = expire
	c.token = res.Header.Get(authTokenHeader)
	if blank(c.token) {
		return ErrorAuth
	}
	c.storageURL, err = url.Parse(res.Header.Get(storageURLHeader))
	if err != nil || blank(c.storageURL.String()) {
		return ErrorAuth
	}

	c.user, c.key = user, key
	now := time.Now()
	c.expireFrom = &now

	return nil
}

// Expired returns true if token is expired or does not exist
func (c *Client) Expired() bool {
	if c.expireFrom == nil || blank(c.token) {
		return true
	}
	duration := time.Duration(c.tokenExpire) * time.Second
	expiredFrom := c.expireFrom.Add(duration).Add(tokenDurationAdd)
	return expiredFrom.Before(time.Now())
}
