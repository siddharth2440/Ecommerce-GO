package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/ecommerce/domain"
	"github.com/golang/ecommerce/services"
)

type Cart_Handler_Struct struct {
	service services.CartService
}

func New_Cart_Handler(service services.CartService) *Cart_Handler_Struct {
	return &Cart_Handler_Struct{
		service: service,
	}
}

// NCh :- New Cart Handler
func (NCh *Cart_Handler_Struct) Create_Cart_Handler(ctx *gin.Context) {
	var cart domain.Cart
	userId := ctx.GetString("userId")

	if err := ctx.ShouldBindJSON(&cart); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"success": false,
				"err":     fmt.Errorf("error in geting the cart data"),
			},
		)
	}

	cartChan := make(chan *domain.Cart, 32)
	errChan := make(chan error, 32)

	go func() {
		cart, err := NCh.service.Create_Cart_Service(&cart, userId)
		if err != nil {
			errChan <- err
			return
		}
		cartChan <- cart
	}()

	for {
		select {
		case cart := <-cartChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"cart":    cart,
			})
			return
		case err := <-errChan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"err":     err.Error(),
			})
			return
		}
	}

}

func (NCh *Cart_Handler_Struct) Get_My_Cart(ctx *gin.Context) {
	userId := ctx.GetString("userId")

	cartChan := make(chan *domain.Cart, 32)
	errChan := make(chan error, 32)

	go func() {
		cart, err := NCh.service.Get_Cart_Details(userId)
		if err != nil {
			errChan <- err
			return
		}
		cartChan <- cart
	}()

	for {
		select {
		case cart := <-cartChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"cart":    cart,
			})
			return
		case err := <-errChan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"err":     err.Error(),
			})
			return
		}
	}

}
