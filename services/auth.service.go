package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang/ecommerce/domain"
	"github.com/golang/ecommerce/response"
	"github.com/golang/ecommerce/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Auth_Service_Struct struct {
	db *mongo.Client
}

// Interfaces Signup login logout

func NewAuthService(Db *mongo.Client) *Auth_Service_Struct {
	return &Auth_Service_Struct{
		db: Db,
	}
}

// NAs := New Auth Servicew

// Signup handler
func (NAs *Auth_Service_Struct) Sign_Up_Service(user *domain.User) (error, *domain.User) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	if user.Username == "" || user.Email == "" || user.Gender == "" || user.Password == "" {
		return errors.New("Missing required fields"), nil
	}

	// Format our Data
	newUser := domain.NewUser(&user.Username, &user.Email, &user.Gender, &user.Password)

	fmt.Println(newUser)
	// var _res_user domain.User
	errChan := make(chan error, 32)
	userChan := make(chan domain.User, 32)
	fmt.Println("Username")
	fmt.Println(user.Username)
	// find the User i.e., is that user exist or not
	// to_find_the_user := bson.M{
	// 	"username": user.Username,
	// }

	go func() {
		defer func() {
			close(errChan)
			close(userChan)
		}()

		// err := NAs.db.Database("ecommerce_golang").Collection("users").FindOne(ctx, to_find_the_user).Decode(&newUser)

		// if err != nil {
		// 	errChan <- err
		// 	return
		// }

		// fmt.Println(newUser.Email)

		// if newUser.Username != "" {
		// 	errChan <- fmt.Errorf("user already exists")
		// 	return
		// }

		// It means that User is not exists in our Database So we need to create a user
		insert_res, err := NAs.db.Database("ecommerce_golang").Collection("users").InsertOne(ctx, newUser)
		if err != nil {
			errChan <- err
			return
		}

		fmt.Println(insert_res)
		userChan <- *newUser

	}()
	select {
	case err := <-errChan:
		return err, nil
	case res_user := <-userChan:
		return nil, &res_user
	case <-ctx.Done():
		return context.DeadlineExceeded, nil
	}
}

// Login Handler
func (NAs *Auth_Service_Struct) Login_service(login_payload response.LoginResponse) (*domain.User, error) {

	// fmt.Println("Login Service is called")
	// fmt.Println(login_payload)'

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	to_find_user := bson.M{
		"username": login_payload.Username,
	}
	var user domain.User

	loggedIn_user_chan := make(chan *domain.User, 32)
	err_chan := make(chan error, 32)

	if login_payload.Password == "" || login_payload.Username == "" {
		return nil, errors.New("invalid credentials")
	}

	go func() {
		defer func() {
			close(loggedIn_user_chan)
			close(err_chan)
		}()

		fmt.Println("inside the GoRoutine......")
		err := NAs.db.Database("ecommerce_golang").Collection("users").FindOne(ctx, to_find_user).Decode(&user)
		if err != nil {
			err_chan <- err
			return
		}
		fmt.Printf("user Password %v\n", user.Password)
		fmt.Printf("login_payload Password %v\n", login_payload.Password)
		isValidPassword := utils.VerifyPassword(login_payload.Password, user.Password)
		if !isValidPassword {
			err_chan <- fmt.Errorf("invalid password")
			return
		}
		fmt.Println(user)
		loggedIn_user_chan <- &user
	}()

	select {
	case err := <-err_chan:
		return nil, err
	case user_details := <-loggedIn_user_chan:
		return user_details, nil
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}

}

// Logout Hanler
