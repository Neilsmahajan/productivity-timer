package auth

import (
	"encoding/json"
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

type Service interface {
	GetUserFromSession(r *http.Request) (*goth.User, error)
	StoreUserInSession(w http.ResponseWriter, r *http.Request, user *goth.User) error
	ClearUserSession(w http.ResponseWriter, r *http.Request) error
}

type service struct{}

func NewAuth() Service {
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

	return &service{}
}

func (s *service) StoreUserInSession(w http.ResponseWriter, r *http.Request, user *goth.User) error {
	session, err := gothic.Store.Get(r, "user-session")
	if err != nil {
		return err
	}

	// Store the marshaled user data
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	session.Values["user"] = string(userJSON)
	return session.Save(r, w)
}

func (s *service) ClearUserSession(w http.ResponseWriter, r *http.Request) error {
	session, err := gothic.Store.Get(r, "user-session")
	if err != nil {
		return err
	}

	// Clear the session by setting MaxAge to -1
	session.Options.MaxAge = -1
	delete(session.Values, "user")

	return session.Save(r, w)
}

func (s *service) GetUserFromSession(r *http.Request) (*goth.User, error) {
	// Get the session
	session, err := gothic.Store.Get(r, "user-session")
	if err != nil {
		return nil, err
	}

	userValue := session.Values["user"]
	if userValue == nil {
		return nil, fmt.Errorf("no user in session")
	}

	// Unmarshal the user data
	userString, ok := userValue.(string)
	if !ok {
		return nil, fmt.Errorf("invalid session data")
	}

	var user goth.User
	if err = json.Unmarshal([]byte(userString), &user); err != nil {
		return nil, err
	}

	return &user, nil
}
