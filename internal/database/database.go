package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/neilsmahajan/productivity-timer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service interface {
	Health() map[string]string
	FindOrCreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	UpsertTimerSession(ctx context.Context, timerSession *models.TimerSession) error
	GetTimerSessionByID(ctx context.Context, userId, tag string) (*models.TimerSession, error)
}

type service struct {
	db *mongo.Client
}

var (
	host     = os.Getenv("DB_HOST")
	port     = os.Getenv("DB_PORT")
	database = os.Getenv("DB_DATABASE")
	username = os.Getenv("DB_USERNAME")
	password = os.Getenv("DB_ROOT_PASSWORD")
)

func New() Service {
	var uri string
	if username != "" && password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%s", username, password, host, port)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%s", host, port)
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	return &service{
		db: client,
	}
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.db.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("db down: %v", err)
	}

	return map[string]string{
		"message": "It's healthy",
	}
}

func (s *service) getUsersCollection() *mongo.Collection {
	return s.db.Database(database).Collection("users")
}

func (s *service) getTimerSessionsCollection() *mongo.Collection {
	return s.db.Database(database).Collection("timers")
}

func (s *service) FindOrCreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	collection := s.getUsersCollection()

	// Try to find existing user by provider and provider ID
	filter := bson.M{
		"provider":    user.Provider,
		"provider_id": user.ProviderID,
	}

	var existingUser models.User
	err := collection.FindOne(ctx, filter).Decode(&existingUser)

	if err == nil {
		// User exists, update last login
		update := bson.M{
			"$set": bson.M{
				"last_login_at": time.Now(),
				"email":         user.Email,
				"name":          user.Name,
				"first_name":    user.FirstName,
				"last_name":     user.LastName,
				"nick_name":     user.NickName,
				"avatar_url":    user.AvatarURL,
			},
		}
		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		existingUser.LastLoginAt = time.Now()
		return &existingUser, nil
	}

	if err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// User doesn't exist, create new one
	now := time.Now()
	user.CreatedAt = now
	user.LastLoginAt = now

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *service) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	collection := s.getUsersCollection()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *service) UpsertTimerSession(ctx context.Context, timerSession *models.TimerSession) error {
	collection := s.getTimerSessionsCollection()
	filter := bson.M{"_id": timerSession.ID}

	err := collection.FindOne(ctx, filter).Decode(&timerSession)
	if err == nil {
		update := bson.M{
			"$set": bson.M{
				"last_login_at": time.Now(),
				"duration":      timerSession.StartTime.Sub(time.Now()),
			},
		}
		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			return fmt.Errorf("failed to update timer session: %w", err)
		}
		return nil
	}

	if err != mongo.ErrNoDocuments {
		return fmt.Errorf("database error: %w", err)
	}
	_, err = collection.InsertOne(ctx, timerSession)
	if err != nil {
		return fmt.Errorf("failed to insert timer session: %w", err)
	}

	return nil
}

func (s *service) GetTimerSessionByID(ctx context.Context, userId, tag string) (*models.TimerSession, error) {
	collection := s.getTimerSessionsCollection()
	var timerSession models.TimerSession
	err := collection.FindOne(ctx, bson.M{"user_id": userId, "tag": tag}).Decode(&timerSession)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
	}
	timerSession.Duration = (int64)(time.Now().Sub(timerSession.StartTime).Seconds())
	return &timerSession, nil
}
