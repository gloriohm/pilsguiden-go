package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type CtxKey string

const sessionCookieName = "session_id"

var sessionKey = CtxKey("session_id")

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sessID string

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

		ctx := context.WithValue(r.Context(), sessionKey, sessID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateSessionID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63())
}

func GetSessionID(r *http.Request) string {
	val := r.Context().Value(sessionKey)
	if id, ok := val.(string); ok {
		return id
	}
	return ""
}
