package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessCookie, err := r.Cookie("access_token")
		if err != nil {
			log.Println("[AUTH] Missing access_token cookie")
			redirectToLogin(w, r)
			return
		}

		accessToken := accessCookie.Value
		token, err := jwt.Parse(accessToken, JwtKeys.Keyfunc)
		if err != nil || !token.Valid {
			log.Printf("[AUTH] Access token invalid: %v", err)

			if r.URL.Query().Get("refreshed") == "1" {
				log.Println("[AUTH] Already tried refreshing token, giving up.")
				redirectToLogin(w, r)
				return
			}

			if tryRefreshToken(w, r) {
				newURL := *r.URL
				q := newURL.Query()
				q.Set("refreshed", "1")
				newURL.RawQuery = q.Encode()

				log.Println("[AUTH] Access token refreshed")
				http.Redirect(w, r, newURL.String(), http.StatusTemporaryRedirect)
				return
			}

			redirectToLogin(w, r)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			redirectToLogin(w, r)
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			redirectToLogin(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	log.Printf("[AUTH] Redirecting to /login from %s\n", r.URL.Path)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func tryRefreshToken(w http.ResponseWriter, r *http.Request) bool {
	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil || refreshCookie.Value == "" {
		log.Println("[AUTH] No refresh token found in cookies")
		return false
	}

	// Prepare request body
	payload := map[string]string{
		"refresh_token": refreshCookie.Value,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(
		"POST",
		os.Getenv("SUPABASE_URL")+"/auth/v1/token?grant_type=refresh_token",
		bytes.NewReader(body),
	)
	if err != nil {
		log.Printf("[AUTH] Failed to create refresh request: %v", err)
		return false
	}

	req.Header.Set("apikey", os.Getenv("SUPABASE_PUBLIC_KEY"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("[AUTH] Token refresh failed: %v", err)
		if resp != nil {
			log.Printf("[AUTH] Status: %d", resp.StatusCode)
			respBody, _ := io.ReadAll(resp.Body)
			log.Printf("[AUTH] Response: %s", respBody)
		}
		return false
	}
	defer resp.Body.Close()

	var res struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"` // seconds
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		log.Printf("[AUTH] Failed to parse refresh response: %v", err)
		return false
	}

	// Set new cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    res.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(time.Duration(res.ExpiresIn) * time.Second),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour), // or whatever is reasonable
	})

	return true
}
