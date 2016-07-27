package handlers

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"pixur.org/pixur/status"
)

const (
	xsrfCookieName    = "XSRF-TOKEN" // From angular
	xsrfHeaderName    = "X-XSRF-TOKEN"
	xsrfTokenLength   = 128 / 8
	xsrfTokenLifetime = time.Hour * 24 * 365 * 10
)

var (
	b64XsrfEnc         = base64.RawStdEncoding
	b64XsrfTokenLength = b64XsrfEnc.EncodedLen(xsrfTokenLength)
)

var (
	random io.Reader        = rand.Reader
	now    func() time.Time = time.Now
)

type xsrfCookieKey struct{}
type xsrfHeaderKey struct{}

func newXsrfToken(random io.Reader) (string, error) {
	xsrfToken := make([]byte, xsrfTokenLength)
	if _, err := io.ReadFull(random, xsrfToken); err != nil {
		return "", status.InternalError(err, "can't create xsrf token")
	}

	b64XsrfToken := make([]byte, b64XsrfTokenLength)
	b64XsrfEnc.Encode(b64XsrfToken, xsrfToken)
	return string(b64XsrfToken), nil
}

func newXsrfCookie(token string, now func() time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     xsrfCookieName,
		Value:    token,
		Path:     "/", // Has to be accessible from root javascript, reset from previous
		Expires:  now().Add(xsrfTokenLifetime),
		Secure:   true,
		HttpOnly: false,
	}
}

// newXsrfContext adds the cookie and header xsrf tokens to ctx
func newXsrfContext(ctx context.Context, cookie, header string) context.Context {
	ctx = context.WithValue(ctx, xsrfCookieKey{}, cookie)
	return context.WithValue(ctx, xsrfHeaderKey{}, header)
}

// fromXsrfContext extracts the cookie and header xsrf tokens from ctx
func fromXsrfContext(ctx context.Context) (cookie string, header string, ok bool) {
	c := ctx.Value(xsrfCookieKey{})
	h := ctx.Value(xsrfHeaderKey{})
	if c != nil && h != nil {
		return c.(string), h.(string), true
	}
	return "", "", false
}

// fromXsrfRequest extracts the cookie and header xsrf tokens from r
func fromXsrfRequest(r *http.Request) (cookie string, header string, err error) {
	c, err := r.Cookie(xsrfCookieName)
	if err == http.ErrNoCookie {
		return "", "", status.Unauthenticated(err, "missing xsrf cookie")
	} else if err != nil {
		// this can't happen according to the http docs
		return "", "", status.InternalError(err, "can't get xsrf token from cookie")
	}
	h := r.Header.Get(xsrfHeaderName)
	return c.Value, h, nil
}

// checkXsrfContext extracts the xsrf tokens and make sure they match
func checkXsrfContext(ctx context.Context) error {
	c, h, ok := fromXsrfContext(ctx)
	if !ok {
		return status.Unauthenticated(nil, "missing xsrf token")
	}
	// check the encoded length, not the binary length
	if len(c) != b64XsrfTokenLength {
		return status.Unauthenticated(nil, "wrong length xsrf token")
	}
	if subtle.ConstantTimeCompare([]byte(h), []byte(c)) != 1 {
		return status.Unauthenticated(nil, "xsrf tokens don't match")
	}
	return nil
}