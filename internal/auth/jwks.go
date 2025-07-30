package auth

import (
	"context"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
)

var JwtKeys keyfunc.Keyfunc

func init() {
	var url []string
	url = append(url, "https://yumsesvtkvukjaqqzgwz.supabase.co/auth/v1/.well-known/jwks.json")
	ctx := context.Background()

	var err error
	JwtKeys, err = keyfunc.NewDefaultCtx(ctx, url)
	if err != nil {
		panic(fmt.Sprintf("Failed to load JWKS: %v", err))
	}
}
