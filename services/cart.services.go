package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/golang/ecommerce/domain"
	"github.com/golang/ecommerce/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CartService interface {
	Create_Cart_Service(cart *domain.Cart, userID string) (*domain.Cart, error)
	Get_Cart_Details(userId string) (*domain.Cart, error)
	Delete_Cart() (*domain.Cart, error)
	Get_All_Carts() (*[]domain.Cart, error)
	Update_Cart() (*[]domain.Cart, error)
}

type Cart_Service_Struct struct {
	db *mongo.Client
}

func New_Cart_Service(db *mongo.Client) *Cart_Service_Struct {
	return &Cart_Service_Struct{
		db: db,
	}
}

// NCs := New Cart Services
// create cart
func (NCs *Cart_Service_Struct) Create_Cart_Service(cart *domain.Cart, userID string) (*domain.Cart, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	cartChan := make(chan *domain.Cart, 32)
	errChan := make(chan error, 32)

	redis_client := utils.Get_Redis()
	if redis_client == nil {
		errChan <- fmt.Errorf("error in connecting with redis client")
	}
	user_objId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		errChan <- err
	}

	new_cart := domain.NewCart(&(*cart).Products, &user_objId)
	var wg sync.WaitGroup
	wg.Add(2)

	// Set it to the Redis
	go func() {
		defer func() {
			wg.Done()
		}()
		// Set the cart details in the Redis
		_, err := redis_client.HExists("cart:"+userID, "cart").Result()
		if err != nil {
			errChan <- err
			return
		}
		cart_json, err := json.Marshal(new_cart)
		if err != nil {
			errChan <- err
			return
		}
		isAdded, err := redis_client.HSet("cart:"+userID, "cart", string(cart_json)).Result()
		if err != nil {
			errChan <- err
			return
		}
		if !isAdded {
			errChan <- fmt.Errorf("error in adding cart to redis")
			return
		}
	}()
	// MongoDB
	go func() {
		defer func() {
			wg.Done()
		}()

		insertResult, err := NCs.db.Database("ecommerce_golang").Collection("carts").InsertOne(ctx, new_cart)
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println("insertResult")
		fmt.Println(insertResult)

		cartChan <- new_cart
	}()

	wg.Wait()

	for {
		select {
		case cart := <-cartChan:
			return cart, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, fmt.Errorf("context deadline exceeded")
		}
	}
}

// get Cart
func (NCs *Cart_Service_Struct) Get_Cart_Details(userId string) (*domain.Cart, error) {
	// from Redis

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	cartChan := make(chan *domain.Cart, 32)
	errChan := make(chan error, 32)
	var cartData domain.Cart

	redis_client := utils.Get_Redis()
	if redis_client == nil {
		return nil, fmt.Errorf("error in connecting with redis client")
	}

	go func() {
		fmt.Println("Getting Data from Redis...")
		red_res, err := redis_client.HGet("cart:"+userId, "cart").Result()
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println("red_res")
		fmt.Println(red_res)

		err = json.Unmarshal([]byte(red_res), &cartData)
		if err != nil {
			errChan <- err
			return
		}
		cartChan <- &cartData
	}()
	for {
		select {
		case cart := <-cartChan:
			return cart, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, fmt.Errorf("context deadline exceeded")
		}
	}
}

// delete Cart
func (NCs *Cart_Service_Struct) Delete_Cart() (*domain.Cart, error) {
	return nil, nil
}

// Get All carts
func (NCs *Cart_Service_Struct) Get_All_Carts() (*[]domain.Cart, error) {
	return nil, nil
}

func (NCs *Cart_Service_Struct) Update_Cart() (*[]domain.Cart, error) {
	return nil, nil
}
