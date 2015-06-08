package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"hash"
	"net/http"
	"time"
)

var (
	ErrNoAuthentication = errors.New("No authentication scheme was configured.")
	ErrEmptyAuth        = errors.New("Empty or missing authentication header.")
	ErrNoGrantTime      = errors.New("No grant time specified in the authentication grant.")
	ErrExpiredGrant     = errors.New("The authentication grant has expired.")
	ErrMACMismatch      = errors.New("The provided message authentication code is invalid for the given message.")
)

type AuthenticatedUser struct {
	UserID               int64     `json:"user_id"`
	GrantTime            time.Time `json:"grant_time"`
	GrantDurationSeconds int64     `json:"grant_duration_sec"`
}

type Authenticator interface {
	GetUser(*http.Request) (*AuthenticatedUser, error)
}

type PassthroughAuthenticator struct{}

func (auth *PassthroughAuthenticator) GetUser(req *http.Request) (*AuthenticatedUser, error) {
	return nil, ErrNoAuthentication
}

type HMACAuthenticator struct {
	key []byte
	h   func() hash.Hash
	now time.Time
}

func (auth *HMACAuthenticator) SetTime(t time.Time) {
	auth.now = t
}

func NewHMACAuthenticatorSHA256(key []byte) *HMACAuthenticator {
	return &HMACAuthenticator{
		key: key,
		h:   sha256.New,
	}
}

func (auth *HMACAuthenticator) GetUser(req *http.Request) (*AuthenticatedUser, error) {
	authHeader := []byte(req.Header.Get("Authorization"))
	userProvidedHmacBase64 := req.Header.Get("X-Authorization-HMAC")

	if len(authHeader) == 0 || userProvidedHmacBase64 == "" {
		return nil, ErrEmptyAuth
	}

	userProvidedHmac, _ := base64.StdEncoding.DecodeString(userProvidedHmacBase64)

	macWriter := hmac.New(auth.h, auth.key)
	macWriter.Write(authHeader)
	expectedMac := macWriter.Sum(nil)

	if hmac.Equal(expectedMac, userProvidedHmac) {
		var authUser AuthenticatedUser
		err := json.Unmarshal(authHeader, &authUser)
		// Valid JSON but no shared values will unmarshal to the zero valued authenticated user; only pass back
		// a non-zero-valued authenticated user
		if err == nil && authUser.UserID != 0 {
			if authUser.GrantTime.IsZero() {
				return nil, ErrNoGrantTime
			} else if authUser.GrantTime.Add(time.Duration(authUser.GrantDurationSeconds) * time.Second).Before(auth.now) {
				return nil, ErrExpiredGrant
			} else {
				return &authUser, nil
			}
		}
	}

	return nil, ErrMACMismatch
}
