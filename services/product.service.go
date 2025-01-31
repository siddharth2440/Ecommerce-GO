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
func (NPs *Product_Service_Struct) Get_Latest_Products() (*[]domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	productsChan := make(chan *[]domain.Product, 32)
	errChan := make(chan error, 32)

	redis_cient := utils.Get_Redis()
	if redis_cient == nil {
		errChan <- fmt.Errorf("error in connecting with redis client")
	}

	var products []domain.Product

	to_get_latest_products := bson.A{
		bson.M{
			"$sort": bson.M{
				"createdAt": -1,
			},
		},
		bson.M{
			"$limit": 2,
		},
	}

	go func() {
		// search in redis
		num_of_proucts, err := redis_cient.LLen("products").Result()
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println("num_of_products")
		fmt.Println(num_of_proucts)
		if num_of_proucts < 2 {
			fmt.Println("Get Latest Products from Mongodb")
			// get data from database mongoDb
			cur, err := NPs.db.Database("ecommerce_golang").Collection("products").Aggregate(ctx, to_get_latest_products)
			if err != nil {
				errChan <- err
				return
			}

			for cur.Next(ctx) {
				var prod domain.Product
				err := cur.Decode(&prod)
				if err != nil {
					errChan <- err
					return
				}
				// fmt.Println("Product:")
				// fmt.Println(prod)
				products = append(products, prod)
				_product, err := json.Marshal(prod)
				if err != nil {
					errChan <- err
					return
				}
				// fmt.Println(products)
				// set in Redis
				res, err := redis_cient.LPush("products", string(_product)).Result()
				if err != nil {
					errChan <- err
				}
				fmt.Printf("LPush Res :- %v\n", res)
			}
			// set expiry for products
			_, err = redis_cient.Expire("products", time.Second*10).Result()
			if err != nil {
				errChan <- err
				return
			}
			fmt.Println("*products")
			fmt.Println(products)
			productsChan <- &products
		} else {

			fmt.Println("Get Latest Products from Redis")
			_products, err := redis_cient.LRange("products", 0, 2).Result()
			if err != nil {
				errChan <- err
				return
			}
			for _, product := range _products {
				var prod domain.Product
				err := json.Unmarshal([]byte(product), &prod)
				if err != nil {
					errChan <- err
					return
				}
				products = append(products, prod)
			}
			productsChan <- &products
		}
	}()

	for {
		select {
		case products := <-productsChan:
			return products, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}

}

// Delete the Product
func (NPs *Product_Service_Struct) Delete_Products_Details(product_id, userid *string) (*domain.Product, error) {
	fmt.Printf("Deleting the Product... %s", *product_id)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	productChan := make(chan *domain.Product, 32)
	errChan := make(chan error, 32)

	prod_objId, err := primitive.ObjectIDFromHex(*product_id)
	if err != nil {
		errChan <- err
	}

	to_delete := bson.M{
		"_id": bson.M{
			"$eq": prod_objId,
		},
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// to delete from the Primary DB
	go func() {
		defer wg.Done()
		var prod domain.Product
		err := NPs.db.Database("ecommerce_golang").Collection("products").FindOne(
			ctx,
			bson.M{
				"_id": bson.M{
					"$eq": prod_objId,
				},
			},
		).Decode(&prod)
		if err != nil {
			errChan <- err
			return
		}

		del_res, err := NPs.db.Database("ecommerce_golang").Collection("products").DeleteOne(ctx, to_delete)
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println("Deleted Product:", del_res.DeletedCount)
		productChan <- &prod
	}()
	/// first chk in Redis and then delete from the redis if that product exists in the Redis
	go func() {
		defer wg.Done()
		redis_client := utils.Get_Redis()
		if redis_client == nil {
			errChan <- fmt.Errorf("error in connecting with redis client")
			return
		}
		// chk in redis
		total_Products, err := redis_client.LLen("userProducts:" + *userid).Result()
		if err != nil {
			errChan <- err
			return
		}

		fmt.Println("totalsProducts")
		fmt.Println(total_Products)
		if total_Products > 0 {
			products, err := redis_client.LRange("userProducts:"+*userid, 0, total_Products-1).Result()
			fmt.Println("Products")
			fmt.Println(products)
			if err != nil {
				errChan <- err
				return
			}
			for _, product := range products {
				var prod domain.Product
				err := json.Unmarshal([]byte(product), &prod)
				if err != nil {
					errChan <- err
					return
				}
				// fmt.Println("Iterating over product :-")
				// fmt.Println(prod)
				if prod.ID == prod_objId {
					// fmt.Println("product Id Matches")
					// fmt.Println(prod.ID)
					// delete the product from the redis
					m_prod, err := json.Marshal(prod)
					if err != nil {
						errChan <- err
						return
					}
					// fmt.Println("m_prod")
					// fmt.Println(string(m_prod))

					del_res, err := redis_client.LRem("userProducts:"+*userid, 1, string(m_prod)).Result()
					if err != nil {
						errChan <- err
						return
					}
					// print delete result
					fmt.Println("Delete Prodct Response from Redis")
					fmt.Println(del_res)
				}
			}
		}

	}()

	wg.Wait()

	for {
		select {
		case product := <-productChan:
			return product, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}

}

// Update the product information
func (NPs *Product_Service_Struct) Update_Products_Details(product_id *string, userid *string, update_product *domain.To_update_product) (*domain.Product, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	prodChan := make(chan domain.Product, 32)
	errChan := make(chan error, 32)

	// var product domain.Product

	prod_objId, err := primitive.ObjectIDFromHex(*product_id)
	if err != nil {
		errChan <- err
	}
	// user_objId, err := primitive.ObjectIDFromHex(*userid)
	// if err != nil {
	// 	errChan <- err
	// }

	redis_client := utils.Get_Redis()
	if redis_client != nil {
		errChan <- fmt.Errorf("error in getting the redis client")
	}

	var wg sync.WaitGroup

	wg.Add(2)

	// to update the details of Product in mongodb
	go func() {
		defer wg.Done()
		var prod domain.Product
		// get data from the mongodb
		err := NPs.db.Database("ecommerce_golang").Collection("products").FindOne(ctx, bson.M{
			"_id": bson.M{
				"$eq": prod_objId,
			},
		}).Decode(&prod)
		if err != nil {
			errChan <- err
			return
		}

		if (*update_product).Title != "" {
			prod.Title = (*update_product).Title
		}
		if (*update_product).Desc != "" {
			prod.Desc = (*update_product).Desc
		}
		if (*update_product).Img != "" {
			prod.Img = (*update_product).Img
		}
		if (*update_product).Price != 0 {
			prod.Price = (*update_product).Price
		}
		if len((*update_product).Size) != 0 {
			prod.Size = (*update_product).Size
		}

		updateRes, err := NPs.db.Database("ecommerce_golang").Collection("products").UpdateOne(ctx,
			bson.M{
				"_id": bson.M{
					"$eq": prod_objId,
				},
			},
			bson.M{
				"$set": prod,
			},
		)

		if err != nil {
			errChan <- err
			return
		}
		fmt.Printf("Updated Product: %v", updateRes.ModifiedCount)
		prodChan <- prod
	}()

	// to update the details of Product in redis
	go func() {
		defer wg.Done()

		no_of_products, err := redis_client.LLen("userProducts:" + *userid).Result()
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println("no of products")
		fmt.Println(no_of_products)

		if no_of_products > 0 {
			products, err := redis_client.LRange("userProducts:"+*userid, 0, no_of_products).Result()
			if err != nil {
				errChan <- err
				return
			}

			for _, prod := range products {
				var product domain.Product
				err := json.Unmarshal([]byte(prod), &product)
				if err != nil {
					errChan <- err
					return
				}
				if product.ID == prod_objId {
					fmt.Printf("Matched %v", prod_objId)
					m_prod, err := json.Marshal(product)
					if err != nil {
						errChan <- err
						return
					}
					// remove the element from the Redis
					_, err = redis_client.LRem("userProducts:"+*userid, 1, string(m_prod)).Result()
					if err != nil {
						errChan <- err
						return
					}
					// again add in the redis
					_, err = redis_client.LPush("userProducts:"+*userid, string(m_prod)).Result()
					if err != nil {
						errChan <- err
						return
					}
				}
			}
		}
	}()

	wg.Wait()

	for {
		select {
		case product := <-prodChan:
			return &product, nil
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}

	}
}
