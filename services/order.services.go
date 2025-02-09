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

type OrderService interface {
	Create_Order_Service(order *domain.Order, userId string) (*domain.Order, error)
	Get_User_Orders(userID string) (*domain.Order, error)
	Get_All_Orders() (*[]domain.Order, error)
	Delete_User_Order(userId, orderId string) (*domain.Order, error)
	Update_Order_Details(order *domain.Order, userid, orderid string) (*domain.Order, error)
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

func (NOs *Order_Service_Struct) Delete_User_Order(userId, orderId string) (*domain.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	orderChan := make(chan *domain.Order, 32)
	errChan := make(chan error, 32)

	user_obj_id, _ := primitive.ObjectIDFromHex(userId)
	order_obj_id, _ := primitive.ObjectIDFromHex(orderId)
	fmt.Println("order_obj_id")
	fmt.Println(order_obj_id)

	fmt.Println("user_obj_id")
	fmt.Println(user_obj_id)

	var wg sync.WaitGroup
	wg.Add(2)

	redis_client := utils.Get_Redis()
	if redis_client == nil {
		errChan <- fmt.Errorf("error in connecting with redis client")
	}
	var order domain.Order
	fmt.Printf("UserId %v", userId)
	fmt.Printf("OrderId %v", orderId)

	del_query := bson.M{
		"$and": bson.A{
			bson.M{"userid": user_obj_id},
			bson.M{"_id": order_obj_id},
		},
	}
	// delete Order frrom the Redis
	go func() {
		defer func() {
			wg.Done()
		}()

		fmt.Println("from Redis")

		// chk is the Order Exists or not
		isExists, err := redis_client.HExists("orders:"+userId, "order").Result()
		if err != nil {
			errChan <- err
			return
		}
		if !isExists {
			errChan <- fmt.Errorf("order not found")
			return
		}

		// delete the order from the Redis
		red_res, err := redis_client.HDel("orders:"+userId, "order").Result()
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println("Redis Result")
		fmt.Println(red_res)

	}()
	// delete order from the mongodb
	go func() {
		defer func() {
			wg.Done()
		}()
		fmt.Println("from MongoDB")
		err := NOs.db.Database("ecommerce_golang").Collection("orders").FindOne(ctx, bson.M{
			"_id":    order_obj_id,
			"userid": user_obj_id,
		}).Decode(&order)

		fmt.Println("order")
		fmt.Println(order)
		if err != nil {
			errChan <- err
			return
		}

		_, err = NOs.db.Database("ecommerce_golang").Collection("orders").DeleteOne(ctx, del_query)
		if err != nil {
			errChan <- err
			return
		}
		orderChan <- &order
	}()

	wg.Wait()

	for {
		select {
		case err := <-errChan:
			return nil, err
		case order := <-orderChan:
			return order, nil
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}
}

func (NOs *Order_Service_Struct) Update_Order_Details(u_order *domain.Order, userid, orderid string) (*domain.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	orderChan := make(chan *domain.Order, 32)
	errChan := make(chan error, 32)

	fmt.Printf("order %v", u_order)
	fmt.Printf("userid %v", userid)
	fmt.Printf("orderid %v", orderid)

	redis_client := utils.Get_Redis()
	if redis_client == nil {
		errChan <- fmt.Errorf("error in connecting with redis client")
	}

	var updatedOrder domain.Order
	var wg sync.WaitGroup
	wg.Add(1)

	usr_Obj_Id, _ := primitive.ObjectIDFromHex(userid)
	order_Obj_Id, _ := primitive.ObjectIDFromHex(orderid)

	to_find_query := bson.M{
		"_id":    order_Obj_Id,
		"userid": usr_Obj_Id,
	}
	// Update the User Order Details in Redis
	go func() {
		fmt.Println("from Redis")
		defer func() {
			wg.Done()
		}()

		var red_order domain.Order
		isExists, err := redis_client.HExists("orders:"+userid, "order").Result()
		if err != nil {
			errChan <- err
			return
		}
		if !isExists {
			errChan <- fmt.Errorf("order not found")
			return
		}
		order_res, err := redis_client.HGet("orders:"+userid, "order").Result()
		if err != nil {
			errChan <- err
			return
		}
		err = json.Unmarshal([]byte(order_res), &red_order)
		if err != nil {
			errChan <- err
			return
		}

		red_order.Products = (*u_order).Products
		red_order.Amount = (*u_order).Amount
		red_order.Status = (*u_order).Status
		red_res, err := json.Marshal(red_order)
		if err != nil {
			errChan <- err
			return
		}
		err = redis_client.HSet("orders:"+userid, "order", string(red_res)).Err()
		if err != nil {
			errChan <- err
			return
		}
	}()

	wg.Add(1)
	// Update the User Order Details in MongoDb
	go func() {
		defer func() {
			wg.Done()
		}()
		err := NOs.db.Database("ecommerce_golang").Collection("orders").FindOne(ctx, to_find_query).Decode(&updatedOrder)
		if err != nil {
			errChan <- err
			return
		}

		if len((*u_order).Products) != 0 {
			updatedOrder.Products = (*u_order).Products
		}
		if (*u_order).Amount != 0 {
			updatedOrder.Amount = (*u_order).Amount
		}

		if (*u_order).Status != "" {
			updatedOrder.Status = (*u_order).Status
		}
		_, err = NOs.db.Database("ecommerce_golang").Collection("orders").UpdateOne(ctx, to_find_query, bson.M{
			"$set": updatedOrder,
		})
		if err != nil {
			errChan <- err
			return
		}
		orderChan <- &updatedOrder
	}()

	wg.Wait()

	for {
		select {
		case err := <-errChan:
			return nil, err
		case updatedOrderDetails := <-orderChan:
			return updatedOrderDetails, nil
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}
}
