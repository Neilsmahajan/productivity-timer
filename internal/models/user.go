package models

import (
	"time"

	"github.com/markbates/goth"
)

type User struct {
	ID          string    `bson:"_id" json:"id"`
	Email       string    `bson:"email" json:"email"`
	Name        string    `bson:"name" json:"name"`
	FirstName   string    `bson:"first_name" json:"firstName"`
	LastName    string    `bson:"last_name" json:"lastName"`
	NickName    string    `bson:"nick_name" json:"nickName"`
	AvatarURL   string    `bson:"avatar_url" json:"avatarUrl"`
	Provider    string    `bson:"provider" json:"provider"`
	ProviderID  string    `bson:"provider_id" json:"providerId"`
	CreatedAt   time.Time `bson:"created_at" json:"createdAt"`
	LastLoginAt time.Time `bson:"last_login_at" json:"lastLoginAt"`
}

// FromGothUser creates a User from a goth.User
func FromGothUser(gothUser goth.User) *User {
	return &User{
		ID:         gothUser.UserID,
		Email:      gothUser.Email,
		Name:       gothUser.Name,
		FirstName:  gothUser.FirstName,
		LastName:   gothUser.LastName,
		NickName:   gothUser.NickName,
		AvatarURL:  gothUser.AvatarURL,
		Provider:   gothUser.Provider,
		ProviderID: gothUser.UserID,
	}
}
