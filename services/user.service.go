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
			err = redis_client.Set("user_info"+user.ID.Hex(), to_store_user_in_redis, time.Second*10).Err()
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
func (NUs *User_Service_Struct) Update_My_Profile(to_update_user_data *domain.To_update_user, userId *string) (*domain.User, error) {
	fmt.Println("Updating the UserData...")

	user_chan := make(chan *domain.User, 32)
	err_chan := make(chan error, 32)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// search Id of user in Redis
	redis_client := utils.Get_Redis()
	if redis_client == nil {
		return nil, fmt.Errorf("redis connection failed")
	}

	r_data := redis_client.MGet("login_info:username", "login_info:user_id", "login_info:email", "login_info:isAdmin")
	fmt.Println(r_data.Val()...)

	// Print the r-Data using loop
	for idx, val := range r_data.Val() {
		fmt.Println("Index : ", idx)
		fmt.Println("Value : ", val)
	}

	//  Match the redis userId with the provided userId
	if r_data.Val()[1] != *userId {
		err_chan <- fmt.Errorf("invalid userId pProvided")
	}
	//  Update the UserId in database
	var user *domain.User

	go func() {
		defer func() {
			close(user_chan)
			close(err_chan)
		}()

		convert_user_to_objectId, err := primitive.ObjectIDFromHex(r_data.Val()[1].(string))
		if err != nil {
			err_chan <- err
		}
		fmt.Printf("UserId to Object ID: %v", convert_user_to_objectId)

		// find the user in our dataBase and then update it
		to_find_user_filter := bson.M{
			"_id": convert_user_to_objectId,
		}

		err = NUs.db.Database("ecommerce_golang").Collection("users").FindOne(ctx, to_find_user_filter).Decode(&user)
		if err != nil {
			err_chan <- err
			return
		}
		fmt.Println(*user)

		if (*to_update_user_data).Email != "" {
			(*user).Email = (*to_update_user_data).Email
		}

		if (*to_update_user_data).Gender != "" {
			(*user).Gender = (*to_update_user_data).Gender
		}
		if (*to_update_user_data).Email != "" {
			(*user).Username = (*to_update_user_data).Username
		}

		user.UpdatedAt = time.Now()
		fmt.Println("user")
		fmt.Println(*user)

		// update in the dataavbase
		to_update := bson.M{
			"$set": bson.M{
				"username": (*user).Username,
				"email":    (*user).Email,
				"gender":   (*user).Gender,
			},
		}

		var update_result *domain.User
		err = NUs.db.Database("ecommerce_golang").Collection("users").FindOneAndUpdate(ctx, to_find_user_filter, to_update).Decode(&update_result)
		if err != nil {
			err_chan <- err
			return
		}
		fmt.Println("update_result")
		fmt.Println(*update_result)
		// also update the data in our Redis
		// del_val1 := redis_client.Del("login_info:username")
		// del_val2 := redis_client.Del("login_info:email")
		// fmt.Println("del vals")
		// fmt.Println(del_val1)
		// fmt.Println(del_val2)

		// add the new info to the redis
		redis_client.Set("login_info:username", (*user).Username, 0)
		redis_client.Set("login_info:email", (*user).Email, 0)

		user_chan <- user
	}()

	select {
	case user := <-user_chan:
		return user, nil
	case err := <-err_chan:
		return nil, err
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}

// Delete the UserProfile
func (NUs *User_Service_Struct) Delete_My_Profile(userId string) (*domain.User, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	user_chan := make(chan *domain.User, 32)
	err_chan := make(chan error, 32)
	// find and chk the userid is exis or not
	redis_client := utils.Get_Redis()
	if redis_client == nil {
		err_chan <- fmt.Errorf("redis connection failed")
	}

	get_id_from_redis := redis_client.Get("login_info:user_id").Val()
	if get_id_from_redis == "" {
		err_chan <- fmt.Errorf("invalid userId provided")
	}

	user_Object_id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		err_chan <- err
	}
	var user *domain.User

	var wg sync.WaitGroup
	wg.Add(1)
	// to delete the user details from the redis
	go func() {
		defer func() {
			wg.Done()
		}()

		// delete the detilas from the Redis
		del_res := redis_client.Del("login_info:email", "login_info:isAdmin", "login_info:user_id", "login_info:username")
		fmt.Println("del_res")
		fmt.Println(del_res)
	}()

	wg.Add(1)
	// to delete the user details from the MongoDb
	go func() {
		defer func() {
			wg.Done()
		}()

		fmt.Println("Inside the mongo deletion")
		delete_user_filter := bson.M{
			"_id": user_Object_id,
		}

		fmt.Println("user_Object_id")
		fmt.Println(user_Object_id)

		err := NUs.db.Database("ecommerce_golang").Collection("users").FindOneAndDelete(ctx, delete_user_filter).Decode(&user)
		if err != nil {
			err_chan <- err
			return
		}
		fmt.Println("mongo_del_res")

		user_chan <- user

	}()
	wg.Wait()

	select {
	case user_data := <-user_chan:
		return user_data, nil
	case err := <-err_chan:
		return nil, err
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}

// Get any User Profile    -----******
func (NUs *User_Service_Struct) GET_USR_PROFILE(userID string) (*domain.User, error) {

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	// defer cancel()

	user_chan := make(chan *domain.User, 32)
	err_chan := make(chan error, 32)

	// CHK in Redis
	redisClient := utils.Get_Redis()
	var user *domain.User

	if redisClient == nil {
		err_chan <- fmt.Errorf("redis connection failed")
	}

	redis_result := redisClient.HGet("user_info"+userID, "userInfo")
	fmt.Println("redis_result.Val()")
	fmt.Println(redis_result.Result())

	user_object_id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		err_chan <- err
	}

	to_find_user := bson.M{
		"_id": user_object_id,
	}

	user_is_there := redis_result.Val()
	if user_is_there == "" {
		// Get Data from the Mongodb Database
		go func() {
			fmt.Println("User is not there inside the Redis database")

			fmt.Println("ID to match")
			fmt.Println(user_object_id)
			err := NUs.db.Database("ecommerce_golang").Collection("users").FindOne(ctx, to_find_user).Decode(&user)
			if err != nil {
				fmt.Println("Err while decoding the data")
				fmt.Println(err)
				err_chan <- err
				return
			}
			fmt.Println("user")
			fmt.Println(user)
			fmt.Println((*user).Email)
			fmt.Println((*user).Username)

			// Set to the Redis for some time
			is_redis_updated := redisClient.HSet("user_info"+userID, "userInfo", user)

			// Hmget in Hash
			is_redis_updated_with_hmset := redisClient.HMSet(
				"user_info"+userID,
				map[string]interface{}{
					"userInfo": user,
					"username": user.Username,
				})

			if err := is_redis_updated_with_hmset.Err(); err != nil {
				err_chan <- err
			}

			if err := is_redis_updated.Err(); err != nil {
				err_chan <- err
			}

			user_chan <- user
		}()
	} else {
		fmt.Println("From Redis")

		// hmget  user_info [key] userInfo [Field] "username" [Field]
		h_result := redisClient.HMGet("user_info"+userID, "userInfo", "username").Val()
		fmt.Println("h_result")
		fmt.Println(h_result)

		// get all
		data, err := redisClient.HGetAll("user_info" + userID).Result()
		fmt.Println("data")
		fmt.Println(data)

		// exists
		isExists, err := redisClient.HExists("user_info"+userID, "userInfo").Result()
		if err != nil {
			err_chan <- err
		}
		fmt.Println("isExists")
		fmt.Println(isExists)

		// keys
		keys, err := redisClient.HKeys("user_info" + userID).Result()
		if err != nil {
			err_chan <- err
		}
		fmt.Println("keys")
		fmt.Println(keys)

		// values
		values, err := redisClient.HVals("user_info" + userID).Result()
		if err != nil {
			err_chan <- err
		}
		fmt.Println("values")
		fmt.Println(values)

		// Unmarshal the user data to JSON
		err = json.Unmarshal([]byte(user_is_there), &user)
		if err != nil {
			err_chan <- err
		}
		user_chan <- user
	}

	select {
	case user_data := <-user_chan:
		return user_data, nil
	case err := <-err_chan:
		return nil, err
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}

}

// Get Random n no. of users
func (NUs *User_Service_Struct) GET_RANDOM_USERS(userNum int) (*[]domain.User, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	users_chan := make(chan []domain.User, 32)
	err_chan := make(chan error, 32)

	// search in Redis if not exists then  go for MongoDB
	redis_client := utils.Get_Redis()
	var users []domain.User
	if redis_client == nil {
		err_chan <- fmt.Errorf("redis connection failed")
	}

	// Search in Redis
	getUsersFromRedis := redis_client.LLen("random_users").Val()
	fmt.Println("getUsersFromRedis")
	fmt.Println(getUsersFromRedis)

	if getUsersFromRedis > 0 {
		go func() {
			fmt.Println("From Redis")
			getUsersFromRedisArr, err := redis_client.LRange("random_users", 0, int64(userNum)).Result()
			if err != nil {
				err_chan <- err
				return
			}
			fmt.Println("getUsersFromRedisArr")
			fmt.Println(getUsersFromRedisArr)

			for _, user := range getUsersFromRedisArr {
				var get_user domain.User
				err := json.Unmarshal([]byte(user), &get_user)
				if err != nil {
					err_chan <- err
					return
				}
				fmt.Println("Get Usrere user")
				fmt.Println(get_user)
				users = append(users, get_user)
			}
			fmt.Println("*users")
			fmt.Println(users)
			users_chan <- users
		}()
	} else {
		fmt.Println("From MongoDB")

		to_search_random_users := bson.M{
			"$sample": bson.M{
				"size": userNum,
			},
		}

		go func() {
			cur, err := NUs.db.Database("ecommerce_golang").Collection("users").Aggregate(ctx, bson.A{
				to_search_random_users,
			})

			if cur.Err() != nil {
				err_chan <- cur.Err()
				return
			}
			defer cur.Close(ctx)

			if err != nil {
				fmt.Println("Error from get users from the database MongoDB")
				err_chan <- err
				return
			}

			isKeExists, err := redis_client.Exists("random_users").Result()
			if err != nil {
				err_chan <- err
				return
			}

			fmt.Printf("isKeyeEissts Res :- %v\n", isKeExists)

			del_res, err := redis_client.Del("random_users").Result()
			if err != nil {
				err_chan <- err
				return
			}

			fmt.Printf("Del Res :- %v\n", del_res)

			for cur.Next(ctx) {
				var user domain.User
				err := cur.Decode(&user)
				if err != nil {
					err_chan <- err
				}
				json_user, err := json.Marshal(user)
				if err != nil {
					err_chan <- err
				}
				lpush_result, err := redis_client.LPush("random_users", string(json_user)).Result()
				if err != nil {
					err_chan <- err
				}
				fmt.Printf("LPush Res :- %v\n", lpush_result)
				redis_client.LPush("random_users", user)
				(users) = append(users, user)
			}

			// Set TTl for List
			err = redis_client.Expire("random_users", time.Second*10).Err()
			if err != nil {
				err_chan <- err
				return
			}

			users_chan <- users
		}()
	}

	select {
	case users_data := <-users_chan:
		fmt.Println("Data received: ", users_data)
		return &users_data, nil
	case err := <-err_chan:
		return nil, err
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}

// Get Recently joined Users
func (NUs *User_Service_Struct) GET_RECENTLY_JOINED_USERS(userNum int, userId string) (*[]domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	users_chan := make(chan []domain.User, 32)
	err_chan := make(chan error, 32)

	// search in Redis if not exists then  go for MongoDB
	redis_client := utils.Get_Redis()
	var users []domain.User
	if redis_client == nil {
		err_chan <- fmt.Errorf("redis connection failed")
	}

	// Search in Redis
	joined_users, err := redis_client.SCard("recently_joined_users").Result()
	if err != nil {
		err_chan <- fmt.Errorf("getting data from redis is failed: %v", err)
	}
	fmt.Println("joined_users")
	fmt.Println(joined_users)

	user_obj_id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		err_chan <- fmt.Errorf("parsing user id from hex is failed: %v", err)
	}

	if joined_users > 0 {
		go func() {
			fmt.Println("From Redis")
			getUsersFromRedisArr, err := redis_client.SMembers("recently_joined_users").Result()
			if err != nil {
				err_chan <- err
				return
			}
			for _, user := range getUsersFromRedisArr {
				var get_user domain.User
				err := json.Unmarshal([]byte(user), &get_user)
				if err != nil {
					err_chan <- err
					return
				}
				(users) = append(users, get_user)
			}
			users_chan <- users
		}()
	} else {
		fmt.Println("From MongoDB")

		// not me
		// sort the users by createdAt : Desc
		to_find_latest_joined_users := bson.A{
			bson.M{
				"$match": bson.M{
					"_id": bson.M{
						"$ne": user_obj_id,
					},
				},
			},
			bson.M{
				"$sort": bson.M{"createdAt": -1},
			},
			bson.M{
				"$limit": userNum,
			},
		}

		go func() {
			cur, err := NUs.db.Database("ecommerce_golang").Collection("users").Aggregate(ctx, to_find_latest_joined_users)
			if err != nil {
				fmt.Println("Error from get users from the database MongoDB")
				err_chan <- err
				return
			}

			for cur.Next(ctx) {
				var user *domain.User
				err := cur.Decode(&user)
				if err != nil {
					err_chan <- err
				}
				json_user, err := json.Marshal(*user)
				if err != nil {
					err_chan <- err
				}
				if err != nil {
					fmt.Println("Error in Decoding the User: ", err)
					err_chan <- err
					return
				}

				// Adding the user to our set
				res, err := redis_client.SAdd("recently_joined_users", string(json_user)).Result()
				if err != nil {
					err_chan <- fmt.Errorf("error in adding user to redis set: %v", err)
					return
				}
				fmt.Printf("SAdd Res :- %v\n", res)
				(users) = append(users, *user)

				// Setting the TTL for Reently users Set in Redis
				_, err = redis_client.Expire("recently_joined_users", time.Second*6).Result()
				if err != nil {
					err_chan <- fmt.Errorf("error in setting expiration for redis set: %v", err)
					return
				}
				users_chan <- users
			}
		}()
	}

	for {
		select {
		case users_data := <-users_chan:
			return &users_data, nil
		case err := <-err_chan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}
}

// search for user using its username email (not me)
func (NUs *User_Service_Struct) Search_User(query, userId string) (*[]domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	users_chan := make(chan []domain.User, 32)
	err_chan := make(chan error, 32)

	var users []domain.User

	// not me
	// gender, username, email := or
	obj_user_id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		err_chan <- fmt.Errorf("parsing user id from hex is failed: %v", err)
	}

	to_query_user := bson.M{
		"$and": bson.A{
			bson.M{
				"_id": bson.M{
					"$ne": obj_user_id,
				},
			},
			bson.M{
				"$or": bson.A{
					// username
					bson.M{
						"username": bson.M{
							"$regex":   query,
							"$options": "i",
						},
					},
					// email
					bson.M{
						"email": bson.M{
							"$regex":   query,
							"$options": "i",
						},
					},
					// gender
					bson.M{
						"gender": query,
					},
				},
			}},
	}
	// search in Direct mongoDb
	go func() {
		cur, err := NUs.db.Database("ecommerce_golang").Collection("users").Find(ctx, to_query_user)
		if err != nil {
			fmt.Println("Error from get users from the database MongoDB")
			err_chan <- err
			return
		}

		for cur.Next(ctx) {
			var user domain.User
			err := cur.Decode(&user)
			fmt.Printf("User: %v", user)
			if err != nil {
				err_chan <- err
				return
			}
			(users) = append(users, user)
		}
		users_chan <- users
	}()

	for {
		select {
		case users_data := <-users_chan:
			fmt.Printf("Users data: %v", users_data)
			return &users_data, nil
		case err := <-err_chan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}

}
