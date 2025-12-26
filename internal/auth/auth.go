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
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Error saving session: %v", err)
		return err
	}

	log.Printf("User session saved for: %s", user.Email)
	return nil
}

func (s *service) ClearUserSession(w http.ResponseWriter, r *http.Request) error {
	session, err := gothic.Store.Get(r, "user-session")
	if err != nil {
		return err
	}

	// Clear the session by setting MaxAge to -1
	session.Options.MaxAge = -1
	delete(session.Values, "user")

	err = session.Save(r, w)
	if err != nil {
		log.Printf("Error clearing session: %v", err)
		return err
	}

	log.Printf("User session cleared")
	return nil
}

func (s *service) GetUserFromSession(r *http.Request) (*goth.User, error) {
	// Get the session
	session, err := gothic.Store.Get(r, "user-session")
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return nil, err
	}

	// Check if the session has user data
	userValue := session.Values["user"]
	if userValue == nil {
		log.Printf("No user data in session. Session values: %+v", session.Values)
		return nil, fmt.Errorf("no user in session")
	}

	// Unmarshal the user data
	userString, ok := userValue.(string)
	if !ok {
		log.Printf("Invalid session data type: %T", userValue)
		return nil, fmt.Errorf("invalid session data")
	}

	// Parse the user from JSON
	var user goth.User
	err = json.Unmarshal([]byte(userString), &user)
	if err != nil {
		log.Printf("Error unmarshaling user: %v", err)
		return nil, err
	}

	log.Printf("Successfully retrieved user from session: %s", user.Email)
	return &user, nil
}
