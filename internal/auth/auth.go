package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

const (
	maxAge = 86400 * 30
	isProd = false
)

func NewAuth() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	googleClientId := os.Getenv("GOOGLE_KEY")
	googleClientSecret := os.Getenv("GOOGLE_SECRET")
	sessionSecret := os.Getenv("SESSION_SECRET")
	port := os.Getenv("PORT")

	store := sessions.NewCookieStore([]byte(sessionSecret))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = isProd
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store

	callbackURL := fmt.Sprintf("http://localhost:%s/auth/google/callback", port)
	goth.UseProviders(google.New(googleClientId, googleClientSecret, callbackURL))
}
