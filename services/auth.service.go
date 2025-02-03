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

// Interfaces Signup login logout
type AuthService interface {
	Sign_Up_Service(user *domain.User) (*domain.User, error)
	Login_service(login_payload response.LoginResponse) (*domain.User, string, error)
}

type Auth_Service_Struct struct {
	db *mongo.Client
}

func NewAuthService(Db *mongo.Client) *Auth_Service_Struct {
	return &Auth_Service_Struct{
		db: Db,
	}
}

// NAs := New Auth Servicew

// Signup Service
func (NAs *Auth_Service_Struct) Sign_Up_Service(user *domain.User) (*domain.User, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	if user.Username == "" || user.Email == "" || user.Gender == "" || user.Password == "" {
		return nil, errors.New("missing required fields")
	}

	// Format our Data
	newUser := domain.NewUser(&user.Username, &user.Email, &user.Gender, &user.Password)

	fmt.Println(newUser)
	// var _res_user domain.User
	errChan := make(chan error, 32)
	userChan := make(chan domain.User, 32)
	// fmt.Println("Username")
	// fmt.Println(user.Username)
	// find the User i.e., is that user exist or not
	to_find_the_user := bson.M{
		"username": user.Username,
	}

	go func() {
		defer func() {
			close(errChan)
			close(userChan)
		}()

		err := NAs.db.Database("ecommerce_golang").Collection("users").FindOne(ctx, to_find_the_user).Decode(&newUser)

		if err == nil {
			fmt.Println("error nil nhi hai")
			errChan <- fmt.Errorf("user already exists")
			return
		}

		// It means that User is not exists in our Database So we need to create a user
		insert_res, err := NAs.db.Database("ecommerce_golang").Collection("users").InsertOne(ctx, newUser)
		if err != nil {
			errChan <- err
			return
		}

		fmt.Println("insert_res")
		fmt.Println(insert_res)

		userChan <- *newUser

	}()
	select {
	case err := <-errChan:
		return nil, err
	case res_user := <-userChan:
		return &res_user, nil
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}

// Login Handler
func (NAs *Auth_Service_Struct) Login_service(login_payload response.LoginResponse) (*domain.User, string, error) {

	// fmt.Println("Login Service is called")
	// fmt.Println(login_payload)'

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	to_find_user := bson.M{
		"username": login_payload.Username,
	}
	var user domain.User

	loggedIn_user_chan := make(chan *domain.User, 32)
	err_chan := make(chan error, 32)

	if login_payload.Password == "" || login_payload.Username == "" {
		return nil, "", errors.New("invalid credentials")
	}

	go func() {
		defer func() {
			close(loggedIn_user_chan)
			close(err_chan)
		}()

		fmt.Println("inside the GoRoutine......")
		// Store inside the Redis

		// DataBse ddata
		err := NAs.db.Database("ecommerce_golang").Collection("users").FindOne(ctx, to_find_user).Decode(&user)
		if err != nil {
			err_chan <- err
			return
		}
		fmt.Printf("user Password %v\n", user.Password)
		fmt.Printf("login_payload Password %v\n", login_payload.Password)
		isValidPassword := utils.VerifyPassword(login_payload.Password, user.Password)
		if !isValidPassword {
			err_chan <- fmt.Errorf("invalid credentials")
			return
		}
		redis_client := utils.Get_Redis()

		/// Set the Login Data inside the Redis
		red_res_1, err := redis_client.Set("login_info:username", user.Username, 0).Result()
		if err != nil {
			err_chan <- err
			return
		}
		red_res_2, err := redis_client.Set("login_info:user_id", user.ID.Hex(), 0).Result()
		if err != nil {
			err_chan <- err
			return
		}
		red_res_3, err := redis_client.Set("login_info:email", user.Email, 0).Result()
		if err != nil {
			err_chan <- err
			return
		}
		red_res_4, err := redis_client.Set("login_info:isAdmin", user.IsAdmin, 0).Result()
		fmt.Println("red_res")
		fmt.Println(red_res_1)
		fmt.Println(red_res_2)
		fmt.Println(red_res_3)
		fmt.Println(red_res_4)

		if err != nil {
			err_chan <- err
			return
		}

		redis_login_result := redis_client.Get("login_info")
		fmt.Println("redis_login_result")
		fmt.Println(redis_login_result)
		// fmt.Println(user)
		loggedIn_user_chan <- &user
	}()

	select {
	case err := <-err_chan:
		return nil, "", err
	case user_details := <-loggedIn_user_chan:
		// creating A JWT Token
		token, err := utils.Create_JWT_Token(user.ID.Hex(), user.Username, user.IsAdmin)
		if err != nil {
			return nil, "", err
		}
		return user_details, token, nil
	case <-ctx.Done():
		return nil, "", context.DeadlineExceeded
	}
}
