package services

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type User_Service_Struct struct {
	db *mongo.Client
}

func New_User_Service(db *mongo.Client) *User_Service_Struct {
	return &User_Service_Struct{
		db: db,
	}
}

// Get - User - Profile
// NUs :- New User Service
func (NUs *User_Service_Struct) Get_My_Profile(userId string) (string, error) {
	// redis_client := utils.Get_Redis()
	return userId, nil
}
