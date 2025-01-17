package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/ecommerce/domain"
	"github.com/golang/ecommerce/response"
	"github.com/golang/ecommerce/services"
)

type Auth_Interface interface {
}

// Dependency injection of Services
type Auth_Handler_Struct struct {
	services *services.Auth_Service_Struct
}

func New_Auth_Handler(services *services.Auth_Service_Struct) *Auth_Handler_Struct {
	return &Auth_Handler_Struct{
		services: services,
	}
}

// NAh := New_Auth_Handler

// Signup handler
func (NAh *Auth_Handler_Struct) Signup(ctx *gin.Context) {
	var user *domain.User
	fmt.Println("Call Ho rha hai...")

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error":   errors.New(" aa hee nhi rha hai, kuchh toh gadbad hai! "),
			},
		)
	}
	user_chan := make(chan *domain.User, 32)
	err_chan := make(chan error, 32)

	go func() {
		user, err := NAh.services.Sign_Up_Service(user)
		if err != nil {
			err_chan <- err
			return
		}
		user_chan <- user
	}()

	select {
	case err := <-err_chan:
		fmt.Printf("Error in Signup %v\n", err)
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error":   err.Error(),
				"success": false,
			},
		)
	case result_user := <-user_chan:
		ctx.JSON(
			http.StatusAccepted,
			gin.H{
				"success": true,
				"data":    result_user,
			},
		)
	}
}

// Login Handler++
func (NAh *Auth_Handler_Struct) Login(ctx *gin.Context) {

	fmt.Println("Login Handler is called")
	var login_payload response.LoginResponse
	fmt.Println(login_payload)
	if err := ctx.ShouldBindJSON(&login_payload); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error":   errors.New("json me fit nhi ho rha hai"),
			},
		)
	}

	fmt.Println(login_payload)
	// res_chan := make(chan *response.LoginResponse, 32)
	user_chan := make(chan *domain.User, 32)
	err_chan := make(chan error, 32)

	go func() {
		res, token, err := NAh.services.Login_service(login_payload)
		if err != nil {
			err_chan <- err
			return
		}
		fmt.Printf(" Token JWT %s\n ", token)
		ctx.SetCookie("authCookie_golang", token, 3600, "/", "localhost", false, true) // 3600 in seconds
		user_chan <- res
	}()
	select {
	case err := <-err_chan:
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error":   err.Error(),
				"success": false,
			},
		)
	case res_response := <-user_chan:
		ctx.JSON(
			http.StatusAccepted,
			gin.H{
				"success": true,
				"data":    res_response,
			},
		)
	}
}

// Logout Hanler
func (NAh *Auth_Handler_Struct) Logout(ctx *gin.Context) {

	// clear the cookie
	ctx.SetCookie("authCookie_golang", "", -1, "/", "localhost", false, true)

	// and then return from that
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged Out Successfully!",
	})
}
