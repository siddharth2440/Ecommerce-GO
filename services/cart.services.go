package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/golang/ecommerce/domain"
	"github.com/golang/ecommerce/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CartService interface {
	Create_Cart_Service(cart *domain.Cart, userID string) (*domain.Cart, error)
	Get_Cart_Details(userId string) (*domain.Cart, error)
	Delete_Cart(userId, cartId string) (*domain.Cart, error)
	Get_All_Carts() (*[]domain.Cart, error)
	Update_Cart(cart *domain.Cart, userid, cartid string) (*domain.Cart, error)
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
func (NCs *Cart_Service_Struct) Delete_Cart(userId, cartId string) (*domain.Cart, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	cartChan := make(chan *domain.Cart, 32)
	errChan := make(chan error, 32)

	var cart domain.Cart
	redis_client := utils.Get_Redis()
	if redis_client == nil {
		errChan <- fmt.Errorf("unable to connect with Redis server.")
	}

	usr_obj_Id, _ := primitive.ObjectIDFromHex(userId)
	cart_obj_Id, _ := primitive.ObjectIDFromHex(cartId)

	query := bson.M{
		"$and": bson.A{
			bson.M{
				"cart_id": cart_obj_Id,
			},
			bson.M{
				"userid": usr_obj_Id,
			},
		},
	}
	var wg sync.WaitGroup
	wg.Add(2)

	// deleting from redis
	go func() {

		defer wg.Done()
		isExists, err := redis_client.HExists("cart:"+userId, "cart").Result()
		if err != nil {
			errChan <- err
			return
		}
		fmt.Printf("isExists %v\n", isExists)

		if !isExists {
			errChan <- fmt.Errorf("no cart found in Redis for user: %s", userId)
			return
		}
		del_res, err := redis_client.HDel("cart:"+userId, "cart").Result()
		if err != nil {
			errChan <- err
			return
		}

		fmt.Println("Delete Cart Response from Redis")
		fmt.Println(del_res)
	}()

	go func() {
		defer wg.Done()
		err := NCs.db.Database("ecommerce_golang").Collection("carts").FindOneAndDelete(ctx, query).Decode(&cart)
		if err != nil {
			errChan <- err
			return
		}
		cartChan <- &cart
	}()

	wg.Wait()

	for {
		select {
		case cart := <-cartChan:
			return cart, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}
}

// Get All carts
func (NCs *Cart_Service_Struct) Get_All_Carts() (*[]domain.Cart, error) {
	// from MongoDB

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	query_to_get_all_carts := bson.A{
		bson.M{
			"$sort": bson.M{
				"createdAt": -1,
			},
		},
		bson.M{
			"$limit": 3,
		},
	}

	cartsChan := make(chan *[]domain.Cart, 32)
	errChan := make(chan error, 32)
	var carts []domain.Cart

	go func() {
		cur, err := NCs.db.Database("ecommerce_golang").Collection("carts").Aggregate(ctx, query_to_get_all_carts)
		if err != nil {
			errChan <- err
			return
		}
		for cur.Next(ctx) {
			var cart domain.Cart
			err := cur.Decode(&cart)
			if err != nil {
				errChan <- err
				return
			}
			carts = append(carts, cart)
		}

		cartsChan <- &carts
	}()

	for {
		select {
		case carts := <-cartsChan:
			return carts, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}

}

func (NCs *Cart_Service_Struct) Update_Cart(cart *domain.Cart, userid, cartid string) (*domain.Cart, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	cartChan := make(chan *domain.Cart, 32)
	errChan := make(chan error, 32)

	redis_client := utils.Get_Redis()
	if redis_client == nil {
		errChan <- fmt.Errorf("unable to connect with redis server")
	}

	usr_obj_id, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		errChan <- fmt.Errorf("unable to convert userid to its objectI")
	}

	cart_obj_id, _ := primitive.ObjectIDFromHex(cartid)
	filter_query := bson.M{
		"userid":  usr_obj_id,
		"cart_id": cart_obj_id,
	}

	var updatedCart domain.Cart

	var wg sync.WaitGroup
	wg.Add(2)
	// updating in Redis
	go func() {
		_, err := redis_client.HExists("cart:"+userid, "cart").Result()
		if err != nil {
			errChan <- err
			return
		}
		cartByte, err := json.Marshal(*cart)
		if err != nil {
			errChan <- err
			return
		}
		_, err = redis_client.HSet("cart:"+userid, "cart", string(cartByte)).Result()
		if err != nil {
			errChan <- err
			return
		}
		defer wg.Done()
	}()
	// updating in mongodb
	go func() {
		defer wg.Done()
		err := NCs.db.Database("ecommerce_golang").Collection("carts").FindOne(ctx, bson.M{
			"cart_id": cart_obj_id,
		}).Decode(&updatedCart)

		if err != nil {
			errChan <- err
			return
		}
		updatedCart.Products = (*cart).Products
		updatedCart.UserId = cart_obj_id
		updatedCart.CartID = cart_obj_id
		update_query := bson.M{
			"$set": bson.M{
				"products": updatedCart.Products,
			},
		}

		fmt.Println("updatedCart")
		fmt.Println(updatedCart)
		_, err = NCs.db.Database("ecommerce_golang").Collection("carts").UpdateOne(ctx, filter_query, update_query)
		if err != nil {
			errChan <- err
			return
		}
		cartChan <- &updatedCart
	}()
	wg.Wait()

	for {
		select {
		case cart := <-cartChan:
			return cart, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}
}
