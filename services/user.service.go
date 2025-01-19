package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/ecommerce/domain"
	"github.com/golang/ecommerce/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// Get MyProfile
func (NUs *User_Service_Struct) Get_My_Profile(userId string) (*domain.User, error) {

	fmt.Println("Inside the Get_My_profile Method...")
	redis_client := utils.Get_Redis()
	if redis_client == nil {
		return nil, fmt.Errorf("redis connection failed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	userInfo := make(chan domain.User, 32)
	errChan := make(chan error, 32)

	var user domain.User

	// Chk the UserId is valid or not
	user_id := redis_client.Get("login_info:user_id")
	if user_id.Val() != userId {
		errChan <- fmt.Errorf("invalid")
	}

	go func() {
		defer func() {
			close(userInfo)
			close(errChan)
		}()

		// converting the UserID to monogo ObjectID
		obj_Id, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			errChan <- err
		}

		// fmt.Println("obj_Id")
		// fmt.Println(obj_Id)

		to_get_User_details := bson.M{
			"_id": obj_Id,
		}

		// chk in Redis
		user_info_val := redis_client.Get("user_info" + userId).Val()
		fmt.Println("user_info_val")
		fmt.Println(user_info_val)

		if user_info_val == "" {
			fmt.Println("We didn;t get any value")
			err = NUs.db.Database("ecommerce_golang").Collection("users").FindOne(ctx, to_get_User_details).Decode(&user)
			if err != nil {
				errChan <- err
				return
			}

			// Store this information to Redis
			to_store_user_in_redis := &domain.User{
				ID:        user.ID,
				Username:  user.Username,
				Email:     user.Email,
				Gender:    user.Gender,
				IsAdmin:   user.IsAdmin,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			}
			err = redis_client.Set("user_info"+user.ID.Hex(), to_store_user_in_redis, 0).Err()
			if err != nil {
				errChan <- err
			}
			userInfo <- user
		} else {
			got_data_from_redis := redis_client.Get("user_info" + userId)
			fmt.Println("got_data_from_redis")
			fmt.Println(got_data_from_redis.Val())
			// Decode Redis data into user struct
			err = json.Unmarshal([]byte(got_data_from_redis.Val()), &user)
			if err != nil {
				errChan <- err
				return
			}
			userInfo <- user
		}
	}()

	select {
	case err := <-errChan:
		return nil, err
	case user := <-userInfo:
		return &user, nil
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}

// Update the UserProfile

// Delete the UserProfile

// Get any User Profile    -----******

// Get Random n no. of users

// Get Recently joined Users
