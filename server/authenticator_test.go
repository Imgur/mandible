package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestPassthroughAuthenticatorAlwaysReturnsNilUser(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://127.0.0.1/user/123/url", nil)

	authenticator := &PassthroughAuthenticator{}
	user, err := authenticator.GetUser(req)
	if user != nil {
		t.Fatalf("Expected authenticator of the passthrough authenticator to be nil, instead %+v", user)
	}
	if err != ErrNoAuthentication {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
}

func TestHMACAuthenticatorOnValidRequest(t *testing.T) {
	message := AuthenticatedUser{
		UserID:               "123",
		GrantTime:            time.Now(),
		GrantDurationSeconds: 365 * 24 * 3600,
	}
	messageBytes, _ := json.Marshal(&message)
	messageMacWriter := hmac.New(sha256.New, []byte("foobar"))
	messageMacWriter.Write(messageBytes)
	messageMac := base64.StdEncoding.EncodeToString(messageMacWriter.Sum(nil))

	req, _ := http.NewRequest("POST", "http://127.0.0.1/user/123/url", nil)

	req.Header.Set("Authorization", string(messageBytes))
	req.Header.Set("X-Authorization-HMAC", string(messageMac))

	authenticator := NewHMACAuthenticatorSHA256([]byte("foobar"))
	authenticator.SetTime(time.Now())
	user, err := authenticator.GetUser(req)
	if user == nil {
		t.Fatalf("Expected authenticator of of a valid response to not return nil")
	}
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
}

func TestHMACAuthenticatorOnEmptyHeader(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://127.0.0.1/user/123/url", nil)

	req.Header.Set("Authorization", "")

	authenticator := NewHMACAuthenticatorSHA256([]byte("foobar"))
	user, err := authenticator.GetUser(req)
	if user != nil {
		t.Fatalf("Expected authenticator with no auth response to return nil")
	}
	if err != ErrEmptyAuth {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
}

func TestHMACAuthenticatorOnInvalidRequest(t *testing.T) {
	message := AuthenticatedUser{
		UserID:               "123",
		GrantTime:            time.Now(),
		GrantDurationSeconds: 365 * 24 * 3600,
	}
	messageBytes, _ := json.Marshal(&message)
	// wrong key!
	messageMacWriter := hmac.New(sha256.New, []byte("jklfdsjklfsdjklfdsjklfsdjklfsd"))
	messageMacWriter.Write(messageBytes)
	messageMac := base64.StdEncoding.EncodeToString(messageMacWriter.Sum(nil))

	req, _ := http.NewRequest("POST", "http://127.0.0.1/user/123/url", nil)

	req.Header.Set("Authorization", string(messageBytes))
	req.Header.Set("X-Authorization-HMAC", string(messageMac))

	authenticator := NewHMACAuthenticatorSHA256([]byte("foobar"))
	authenticator.SetTime(time.Now())
	user, err := authenticator.GetUser(req)
	if user != nil {
		t.Fatalf("Expected authenticator of of an invalid response to return nil")
	}
	if err != ErrMACMismatch {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
}

func TestHMACAuthenticatorOnExpiredGrant(t *testing.T) {
	grantedTime := time.Now()
	requestTime := time.Now().Add(time.Hour)
	message := AuthenticatedUser{
		UserID:               "123",
		GrantTime:            grantedTime,
		GrantDurationSeconds: 5,
	}
	messageBytes, _ := json.Marshal(&message)
	messageMacWriter := hmac.New(sha256.New, []byte("foobar"))
	messageMacWriter.Write(messageBytes)
	messageMac := base64.StdEncoding.EncodeToString(messageMacWriter.Sum(nil))

	req, _ := http.NewRequest("POST", "http://127.0.0.1/user/123/url", nil)

	req.Header.Set("Authorization", string(messageBytes))
	req.Header.Set("X-Authorization-HMAC", string(messageMac))

	authenticator := NewHMACAuthenticatorSHA256([]byte("foobar"))
	authenticator.SetTime(requestTime)

	user, err := authenticator.GetUser(req)
	if user != nil {
		t.Fatalf("Expected authenticator of of an invalid response to return nil")
	}
	if err != ErrExpiredGrant {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
}
