package domain

import (
	"time"

	"github.com/golang/ecommerce/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id" `
	Username  string             ` json:"username" `
	Email     string             ` json:"email" `
	Gender    string             ` json:"gender" `
	Password  string             ` json:"password" `
	IsAdmin   bool               ` json:"isAdmin" `
	CreatedAt time.Time          ` json:"createdAt" `
	UpdatedAt time.Time          ` json:"UpdatedAt" `
}

func NewUser(username, email, gender, password *string) *User {
	// TODO := Hash the Password i.e., password hashing
	hash, _ := utils.HashPassword(*password)
	return &User{
		ID:        primitive.NewObjectID(),
		Username:  *username,
		Email:     *email,
		Gender:    *gender,
		IsAdmin:   false,
		Password:  hash,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
