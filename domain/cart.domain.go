package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type ProductDetails struct {
	ProductID primitive.ObjectID `json:"ProductID"`
	Quantity  int                `json:"quantity"`
}

type Cart struct {
	ID       primitive.ObjectID `json:"_id" bson:"_id"`
	Products []ProductDetails   `json:"productDetails"`
	UserId   primitive.ObjectID `json:"userId"`
}
