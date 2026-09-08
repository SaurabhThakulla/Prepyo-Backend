package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrGoogleNotConfigured = errors.New("google sign-in is not configured")
	ErrGoogleTokenInvalid  = errors.New("google token is not valid")
	ErrGoogleEmailUnusable = errors.New("google account has no verified email")
)

const googleTokenInfoURL = "https://oauth2.googleapis.com/tokeninfo"

var googleIssuers = map[string]bool{
	"accounts.google.com":         true,
	"https://accounts.google.com": true,
}

type GoogleIdentity struct {
	Subject string
	Email   string
	Name    string
}

type GoogleVerifier struct {
	clientID string
	client   *http.Client
}

func NewGoogleVerifier(clientID string) *GoogleVerifier {
	return &GoogleVerifier{
		clientID: strings.TrimSpace(clientID),
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (v *GoogleVerifier) Configured() bool { return v != nil && v.clientID != "" }

type googleTokenInfo struct {
	Issuer        string `json:"iss"`
	Audience      string `json:"aud"`
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	Expiry        string `json:"exp"`
}

func (v *GoogleVerifier) Verify(ctx context.Context, idToken string) (GoogleIdentity, error) {
	if !v.Configured() {
		return GoogleIdentity{}, ErrGoogleNotConfigured
	}
	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return GoogleIdentity{}, ErrGoogleTokenInvalid
	}

	endpoint := googleTokenInfoURL + "?id_token=" + url.QueryEscape(idToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return GoogleIdentity{}, fmt.Errorf("build tokeninfo request: %w", err)
	}

	res, err := v.client.Do(req)
	if err != nil {
		return GoogleIdentity{}, fmt.Errorf("call tokeninfo: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusBadRequest || res.StatusCode == http.StatusUnauthorized {
		return GoogleIdentity{}, ErrGoogleTokenInvalid
	}
	if res.StatusCode != http.StatusOK {
		return GoogleIdentity{}, fmt.Errorf("tokeninfo returned %d", res.StatusCode)
	}

	var info googleTokenInfo
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return GoogleIdentity{}, fmt.Errorf("decode tokeninfo: %w", err)
	}

	if !googleIssuers[info.Issuer] || info.Audience != v.clientID || info.Subject == "" {
		return GoogleIdentity{}, ErrGoogleTokenInvalid
	}

	expiry, err := strconv.ParseInt(info.Expiry, 10, 64)
	if err != nil || time.Now().After(time.Unix(expiry, 0)) {
		return GoogleIdentity{}, ErrGoogleTokenInvalid
	}

	email := strings.ToLower(strings.TrimSpace(info.Email))
	if email == "" || info.EmailVerified != "true" {
		return GoogleIdentity{}, ErrGoogleEmailUnusable
	}

	name := strings.TrimSpace(info.Name)
	if name == "" {
		name = strings.TrimSpace(info.GivenName)
	}
	if name == "" {
		name, _, _ = strings.Cut(email, "@")
	}

	return GoogleIdentity{Subject: info.Subject, Email: email, Name: name}, nil
}
