package auth

import "errors"

var (
	ErrBindingOauth2Callback = errors.New("unable to bind oauth2 response")
	ErrMissingRequiredScopes = errors.New("missing required scope(s)")
	ErrNotAuthorized         = errors.New("not authorized")
	ErrStateNotSetCorrectly  = errors.New("state not set correctly")
	ErrTokenExchangeFailed   = errors.New("token exchange failed")
)
