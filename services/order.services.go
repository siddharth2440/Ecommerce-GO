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
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderService interface {
	Create_Order_Service(order *domain.Order, userId string) (*domain.Order, error)
	Get_User_Orders(userID string) (*domain.Order, error)
	Get_All_Orders() (*[]domain.Order, error)
}

type Order_Service_Struct struct {
	db *mongo.Client
}

func NewOrderService(db *mongo.Client) *Order_Service_Struct {
	return &Order_Service_Struct{
		db: db,
	}
}

func (NOs *Order_Service_Struct) Create_Order_Service(order *domain.Order, userId string) (*domain.Order, error) {
	order = domain.NewOrder(order)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	orderChan := make(chan *domain.Order, 32)
	errChan := make(chan error, 32)

	redis_client := utils.Get_Redis()
	if redis_client == nil {
		errChan <- fmt.Errorf("error in connecting with redis client")
	}

	var wg sync.WaitGroup
	wg.Add(2)
	// Save Order in Redis
	go func() {
		defer func() {
			wg.Done()
		}()

		_, err := redis_client.HExists("orders:"+userId, "order").Result()
		if err != nil {
			errChan <- err
			return
		}

		m_order, err := json.Marshal(&order)
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println("m_order")
		fmt.Println(string(m_order))
		red_res, err := redis_client.HSet("orders:"+userId, "order", string(m_order)).Result()
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println("Result :")
		fmt.Println(red_res)
	}()
	// Save Order in MongoDB
	go func() {
		defer func() {
			wg.Done()
		}()
		inst_res, err := NOs.db.Database("ecommerce_golang").Collection("orders").InsertOne(ctx, order)
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println("inst_res")
		fmt.Println(inst_res)
		orderChan <- order
	}()

	wg.Wait()
	for {
		select {
		case err := <-errChan:
			return nil, err
		case order := <-orderChan:
			return order, nil
		default:
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func (NOs *Order_Service_Struct) Get_User_Orders(userID string) (*domain.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	// from Redis
	ordersChan := make(chan *domain.Order)
	errChan := make(chan error, 32)

	redis_client := utils.Get_Redis()
	if redis_client == nil {
		errChan <- fmt.Errorf("error in connecting with redis client")
	}

	go func() {
		order, err := redis_client.HGet("orders:"+userID, "order").Result()
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println("order from redis")
		fmt.Println(order)
		var u_order domain.Order
		err = json.Unmarshal([]byte(order), &u_order)
		if err != nil {
			errChan <- err
			return
		}
		ordersChan <- &u_order
	}()

	for {
		select {
		case err := <-errChan:
			return nil, err
		case order := <-ordersChan:
			return order, nil
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}
}

func (NOs *Order_Service_Struct) Get_All_Orders() (*[]domain.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	ordersChan := make(chan *[]domain.Order)
	errChan := make(chan error, 32)

	var orders []domain.Order
	to_get_orders := bson.A{
		bson.M{
			"$sort": bson.M{
				"createdAt": -1,
			},
		},
		bson.M{
			"$limit": 3,
		},
	}
	// from MongoDB
	go func() {
		cur, err := NOs.db.Database("ecommerce_golang").Collection("orders").Aggregate(ctx, to_get_orders)
		if err != nil {
			errChan <- err
			return
		}
		for cur.Next(ctx) {
			var order domain.Order
			err := cur.Decode(&order)
			if err != nil {
				errChan <- err
				return
			}
			orders = append(orders, order)
		}

		ordersChan <- &orders
	}()

	for {
		select {
		case err := <-errChan:
			return nil, err
		case orders := <-ordersChan:
			return orders, nil
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}
}
