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

func (NCh *Cart_Handler_Struct) Delete_Cart_Handler(ctx *gin.Context) {
	cartID := ctx.Param("cartId")
	userId := ctx.GetString("userId")

	cartDetailsChan := make(chan *domain.Cart, 32)
	errChan := make(chan error, 32)

	go func() {

		cart, err := NCh.service.Delete_Cart(userId, cartID)
		if err != nil {
			errChan <- err
			return
		}
		cartDetailsChan <- cart
	}()

	for {
		select {
		case cart := <-cartDetailsChan:
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

func (NCh *Cart_Handler_Struct) Get_Carts_handler(ctx *gin.Context) {

	cartsChan := make(chan *[]domain.Cart, 32)
	errChan := make(chan error, 32)

	go func() {

		carts, err := NCh.service.Get_All_Carts()
		if err != nil {
			errChan <- err
			return
		}
		cartsChan <- carts
	}()

	for {
		select {
		case carts := <-cartsChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"cart":    carts,
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

func (NCh *Cart_Handler_Struct) Update_Cart_handler(ctx *gin.Context) {

	cartChan := make(chan *domain.Cart, 32)
	errChan := make(chan error, 32)

	userId := ctx.GetString("userId")
	cartID := ctx.Param("cartID")
	var new_cart domain.Cart

	if err := ctx.ShouldBindJSON(&new_cart); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"success": false,
				"err":     fmt.Errorf("error in getting the cart data"),
			},
		)
		return
	}
	go func() {

		updated_cart, err := NCh.service.Update_Cart(&new_cart, userId, cartID)
		if err != nil {
			errChan <- err
			return
		}
		cartChan <- updated_cart
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
