package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

type loginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	User         any    `json:"user"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		http.Error(w, "Missing email or password", http.StatusBadRequest)
		return
	}

	payload := loginPayload{Email: email, Password: password}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", os.Getenv("SUPABASE_URL")+"/auth/v1/token?grant_type=password", bytes.NewReader(body))
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		http.Error(w, "Invalid login", http.StatusUnauthorized)
		return
	}

	defer resp.Body.Close()
	var loginRes loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginRes); err != nil {
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    loginRes.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(1 * time.Hour),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    loginRes.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(168 * time.Hour),
	})
}
