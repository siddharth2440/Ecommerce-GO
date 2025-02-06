package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductInfo struct {
	ProductID primitive.ObjectID `json:"product_id" bson:"product_id"`
	Quantity  int                `json:"quantity"`
}

type Order struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Products  []ProductInfo      `json:"products"`
	Amount    int                `json:"amount"`
	UserId    primitive.ObjectID `json:"userId"`
	Address   string             `json:"address" validate:"required,min=2,max=20"`
	Status    string             `json:"status" validate:"required" default:"processing"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

func NewOrder(order *Order) *Order {
	new_order := Order{
		ID:        primitive.NewObjectID(),
		Products:  (*order).Products,
		Amount:    (*order).Amount,
		UserId:    (*order).UserId,
		Address:   (*order).Address,
		Status:    "created",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return &new_order
}
