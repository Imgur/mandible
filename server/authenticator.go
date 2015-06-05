package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"hash"
	"log"
	"net/http"
	"time"
)

type AuthenticatedUser struct {
	UserID               int64     `json:"user_id"`
	GrantTime            time.Time `json:"grant_time"`
	GrantDurationSeconds int64     `json:"grant_duration_sec"`
}

type Authenticator interface {
	GetUser(*http.Request) *AuthenticatedUser
}

type PassthroughAuthenticator struct{}

func (auth *PassthroughAuthenticator) GetUser(req *http.Request) *AuthenticatedUser {
	return nil
}

type HMACAuthenticator struct {
	key []byte
	h   func() hash.Hash
}

func NewHMACAuthenticatorSHA256(key []byte) *HMACAuthenticator {
	return &HMACAuthenticator{
		key: key,
		h:   sha256.New,
	}
}

func (auth *HMACAuthenticator) GetUser(req *http.Request) *AuthenticatedUser {
	authHeader := []byte(req.Header.Get("Authorization"))
	userProvidedHmacBase64 := req.Header.Get("X-Authorization-HMAC")

	if len(authHeader) == 0 || userProvidedHmacBase64 == "" {
		return nil
	}

	userProvidedHmac, _ := base64.StdEncoding.DecodeString(userProvidedHmacBase64)

	macWriter := hmac.New(auth.h, auth.key)
	macWriter.Write(authHeader)
	expectedMac := macWriter.Sum(nil)

	if hmac.Equal(expectedMac, userProvidedHmac) {
		log.Printf("MACs match!")
		var authUser AuthenticatedUser
		err := json.Unmarshal(authHeader, &authUser)
		// Valid JSON but no shared values will unmarshal to the zero valued authenticated user; only pass back
		// a non-zero-valued authenticated user
		if err == nil && authUser.UserID != 0 {
			if authUser.GrantTime.IsZero() {
				log.Printf("Refusing to allow t=0 grant")
				return nil
			} else if authUser.GrantTime.Add(time.Duration(authUser.GrantDurationSeconds) * time.Second).Before(time.Now()) {
				log.Printf("Grant time %s with duration %d has expired!", authUser.GrantTime, authUser.GrantDurationSeconds)
				return nil
			} else {
				return &authUser
			}
		}
	} else {
		log.Printf("Unauthorized! Expected MAC %s but given %s", base64.StdEncoding.EncodeToString(expectedMac), userProvidedHmacBase64)
	}

	return nil
}
