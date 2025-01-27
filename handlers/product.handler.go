package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/ecommerce/domain"
	"github.com/golang/ecommerce/services"
)

type Product_Handler_Struct struct {
	service *services.Product_Service_Struct
}

func New_Product_Handler(service *services.Product_Service_Struct) *Product_Handler_Struct {
	return &Product_Handler_Struct{
		service: service,
	}
}

// NPs := New Product Service
func (NPs *Product_Handler_Struct) Add_Product_Handler(ctx *gin.Context) {
	userId := ctx.GetString("userId")
	var product *domain.Product

	if err := ctx.ShouldBindJSON(&product); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error":   err.Error(),
				"success": false,
			},
		)
	}

	product_chan := make(chan *domain.Product, 32)
	err_chan := make(chan error, 32)

	go func() {
		product, err := NPs.service.Create_Product_Service(userId, product)
		if err != nil {
			err_chan <- err
			return
		}
		product_chan <- product
	}()

	for {
		select {
		case product := <-product_chan:
			ctx.JSON(
				http.StatusOK,
				gin.H{
					"success": true,
					"data":    product,
				})
			return
		case err := <-err_chan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}
	}
}
