package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/ecommerce/domain"
	"github.com/golang/ecommerce/utils"
	"go.mongodb.org/mongo-driver/mongo"
)

type Product_Service_Struct struct {
	db *mongo.Client
}

func NewProductService(db *mongo.Client) *Product_Service_Struct {
	return &Product_Service_Struct{
		db: db,
	}
}

// Create a Product
func (NPs *Product_Service_Struct) Create_Product_Service(userid string, productInfo *domain.Product) (*domain.Product, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	newProduct := domain.NewProduct(&(*productInfo).Title, &(*productInfo).Desc, &(*productInfo).Img, &(*productInfo).Categories, &(*productInfo).Size, &(*productInfo).Color, &(*productInfo).Price, &(*productInfo).InStock)

	product_ch := make(chan domain.Product, 32)
	err_ch := make(chan error, 32)

	// Check in Redis
	redis_client := utils.Get_Redis()

	if redis_client == nil {
		err_ch <- fmt.Errorf("redis connection failed")
	}

	isProducts_exists, err := redis_client.LLen("userProducts:" + userid).Result()
	if err != nil {
		err_ch <- err
	}

	if isProducts_exists >= 10 {
		err_ch <- fmt.Errorf("you have reached the maximum limit of products for this user")
	}

	go func() {
		defer close(product_ch)
		defer close(err_ch)

		// save into the Database
		insert_res, err := NPs.db.Database("ecommerce_golang").Collection("products").InsertOne(ctx, newProduct)
		if err != nil {
			err_ch <- err
			return
		}
		fmt.Println("insert_res")
		fmt.Println(insert_res)
		product, err := json.Marshal(*newProduct)
		if err != nil {
			err_ch <- err
			return
		}
		// save in the redis
		res, err := redis_client.LPush("userProducts:"+userid, string(product)).Result()
		if err != nil {
			err_ch <- err
		}
		fmt.Printf("LPush Res :- %v\n", res)
		product_ch <- *newProduct

	}()

	for {
		select {
		case product := <-product_ch:
			return &product, nil
		case err := <-err_ch:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}
}

// Get latest Products
// Get top 10 Products
