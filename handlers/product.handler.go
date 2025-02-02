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

func (NPh *Product_Handler_Struct) Latest_Products(ctx *gin.Context) {

	productsChan := make(chan *[]domain.Product, 32)
	errChan := make(chan error, 32)

	go func() {
		products, err := NPh.service.Get_Latest_Products()
		if err != nil {
			errChan <- err
			return
		}
		productsChan <- products
	}()

	for {
		select {
		case products := <-productsChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    products,
			})
			return
		case err := <-errChan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}
	}

}

func (NPh *Product_Handler_Struct) Delete_Products(ctx *gin.Context) {
	prodChan := make(chan *domain.Product, 32)
	errChan := make(chan error, 32)

	prod_id := ctx.Param("productId")
	userId := ctx.GetString("userId")

	go func() {
		product, err := NPh.service.Delete_Products_Details(&prod_id, &userId)
		if err != nil {
			errChan <- err
			return
		}
		prodChan <- product
	}()

	for {
		select {
		case product := <-prodChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    product,
			})
			return
		case err := <-errChan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}
	}
}
func (NPh *Product_Handler_Struct) Update_Product_Details(ctx *gin.Context) {

	prod_chan := make(chan *domain.Product, 32)
	err_chan := make(chan error, 32)

	var update_product domain.To_update_product
	if err := ctx.ShouldBindJSON(&update_product); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error":   err.Error(),
				"success": false,
			},
		)
	}

	prod_id := ctx.Param("productId")
	userId := ctx.GetString("userId")

	go func() {
		prod, err := NPh.service.Update_Products_Details(&prod_id, &userId, &update_product)
		if err != nil {
			err_chan <- err
		}
		prod_chan <- prod
	}()

	for {
		select {
		case product := <-prod_chan:
			ctx.JSON(http.StatusOK, gin.H{
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

func (NPh *Product_Handler_Struct) Get_Product_Details_By_ID(ctx *gin.Context) {
	productId := ctx.Param("productId")

	prodChan := make(chan *domain.Product, 32)
	errChan := make(chan error, 32)

	go func() {
		prod, err := NPh.service.Get_Product_Details_By_ID(productId)

		if err != nil {
			errChan <- err
			return
		}
		prodChan <- prod
	}()

	for {
		select {
		case product := <-prodChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    product,
			})
			return
		case err := <-errChan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}
	}
}

func (NPh *Product_Handler_Struct) Product_By_Query(ctx *gin.Context) {
	query := ctx.Query("query")

	prodChan := make(chan *[]domain.Product, 32)
	errChan := make(chan error, 32)

	go func() {
		prods, err := NPh.service.Get_Products_By_Query(query)

		if err != nil {
			errChan <- err
			return
		}
		prodChan <- prods
	}()

	for {
		select {
		case products := <-prodChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    products,
			})
			return
		case err := <-errChan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}
	}
}
