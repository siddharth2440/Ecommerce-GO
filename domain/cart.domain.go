package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type ProductDetails struct {
	ProductID primitive.ObjectID `json:"product_id"`
	Quantity  int                `json:"quantity"`
}

type Cart struct {
	ID       primitive.ObjectID `json:"_id" bson:"_id"`
	CartID   primitive.ObjectID `json:"cart_id" bson:"cart_id"`
	Products []ProductDetails   `json:"productDetails"`
	UserId   primitive.ObjectID `json:"userId"`
}

func NewCart(product *[]ProductDetails, userId *primitive.ObjectID) *Cart {
	return &Cart{
		ID:       primitive.NewObjectID(),
		CartID:   primitive.NewObjectID(),
		Products: *product,
		UserId:   *userId,
	}
}
