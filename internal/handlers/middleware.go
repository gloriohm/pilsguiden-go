package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type RequestData struct {
	Consented bool
	SessionID string
}

type ctxKey struct{}

func WithData(ctx context.Context, data RequestData) context.Context {
	return context.WithValue(ctx, ctxKey{}, data)
}

func GetSessionData(ctx context.Context) RequestData {
	if v, ok := ctx.Value(ctxKey{}).(RequestData); ok {
		return v
	}
	return RequestData{}
}

const sessionCookieName = "session_id"
const consentCookieName = "user_consent"

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sessID string
		var consented bool

		consentCookie, err := r.Cookie(consentCookieName)
		if err != nil || consentCookie.Value == "" {
			consented = true
		}

		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			// New session
			sessID = generateSessionID()
			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookieName,
				Value:    sessID,
				Path:     "/",
				HttpOnly: true,
				Secure:   true, // make sure you're running HTTPS
				SameSite: http.SameSiteLaxMode,
				MaxAge:   1800, // 30 mins
			})
		} else {
			sessID = cookie.Value
		}

		data := RequestData{
			Consented: consented,
			SessionID: sessID,
		}

		ctx := WithData(r.Context(), data)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateSessionID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63())
}
