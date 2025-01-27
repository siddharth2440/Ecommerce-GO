package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID         primitive.ObjectID `bson:"_id" json:"_id" `
	Title      string             `json:"title" validate:"required" `
	Desc       string             `json:"desc" validate:"required" `
	Img        string             `json:"img" validate:"required" `
	Categories []string           `json:"categories" validate:"required" `
	Size       []string           `json:"size" validate:"required" `
	Color      []string           `json:"color" validate:"required" `
	Price      int                `json:"price" validate:"required" `
	InStock    bool               `json:"instock" validate:"required" `
}

func NewProduct(title *string, description *string, image *string, categories *[]string, size *[]string, color *[]string, price *int, inStock *bool) *Product {
	return &Product{
		ID:         primitive.NewObjectID(),
		Title:      *title,
		Desc:       *description,
		Img:        *image,
		Categories: *categories,
		Size:       *size,
		Color:      *color,
		Price:      *price,
		InStock:    *inStock,
	}
}
