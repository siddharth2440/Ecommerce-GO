package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/ecommerce/domain"
	"github.com/golang/ecommerce/services"
)

type Order_Handler_Struct struct {
	services services.OrderService
}

func NewOrderService(service services.OrderService) *Order_Handler_Struct {
	return &Order_Handler_Struct{
		services: service,
	}
}

func (NOh *Order_Handler_Struct) Create_Order_Handler(ctx *gin.Context) {
	var order domain.Order
	userId := ctx.GetString("userId")
	orderChan := make(chan *domain.Order, 32)
	errChan := make(chan error, 32)
	if err := ctx.ShouldBindJSON(&order); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	go func() {
		new_order, err := NOh.services.Create_Order_Service(&order, userId)
		if err != nil {
			errChan <- err
			return
		}
		orderChan <- new_order
	}()

	for {
		select {
		case order := <-orderChan:
			ctx.JSON(http.StatusCreated, gin.H{
				"success": true,
				"order":   order,
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

func (NOh *Order_Handler_Struct) Get_User_Orders_Handler(ctx *gin.Context) {
	userID := ctx.GetString("userId")
	orderChan := make(chan *domain.Order, 32)
	errChan := make(chan error, 32)

	go func() {
		user_order, err := NOh.services.Get_User_Orders(userID)
		if err != nil {
			errChan <- err
			return
		}
		orderChan <- user_order
	}()

	for {
		select {
		case order := <-orderChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"orders":  order,
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

func (NOh *Order_Handler_Struct) Get_Orders_Handler(ctx *gin.Context) {
	orderChan := make(chan *[]domain.Order, 32)
	errChan := make(chan error, 32)

	go func() {
		orders, err := NOh.services.Get_All_Orders()
		if err != nil {
			errChan <- err
			return
		}
		orderChan <- orders
	}()

	for {
		select {
		case orders := <-orderChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"orders":  orders,
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
