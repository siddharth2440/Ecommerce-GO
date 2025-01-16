package domain

import (
	"time"

	"github.com/golang/ecommerce/utils"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID ` json:"id" `
	Username  string    ` json:"username" `
	Email     string    ` json:"email" `
	Gender    string    ` json:"gender" `
	Password  string    ` json:"password" `
	CreatedAt time.Time ` json:"createdAt" `
	UpdatedAt time.Time ` json:"UpdatedAt" `
}

func NewUser(username, email, gender, password *string) *User {
	// TODO := Hash the Password i.e., password hashing
	hash, _ := utils.HashPassword(*password)
	return &User{
		ID:        uuid.New(),
		Username:  *username,
		Email:     *email,
		Gender:    *gender,
		Password:  hash,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
