package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/ecommerce/domain"
	"github.com/golang/ecommerce/services"
)

type User_Handler_Struct struct {
	service *services.User_Service_Struct
}

func New_User_Handler(service *services.User_Service_Struct) *User_Handler_Struct {
	return &User_Handler_Struct{
		service: service,
	}
}

// Get My Profile
// NUh :- New user Handler Struct
func (NUh *User_Handler_Struct) Get_My_Profile(ctx *gin.Context) {
	uId := ctx.GetString("userId")
	isadmin := ctx.GetBool("isAdmin")
	username := ctx.GetString("username")
	fmt.Println(uId)
	fmt.Println(isadmin)
	fmt.Println(username)

	user_chan := make(chan *domain.User, 32)
	err_chan := make(chan error, 32)

	go func() {
		defer func() {
			close(user_chan)
			close(err_chan)
		}()

		user, err := NUh.service.Get_My_Profile(uId)
		if err != nil {
			err_chan <- err
			return
		}
		user_chan <- user

	}()

	select {
	case user := <-user_chan:
		ctx.JSON(
			http.StatusOK,
			gin.H{"message": "User Profile", "data": user})
	case err := <-err_chan:
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"message": "Error fetching user profile", "error": err.Error()})
	}

}
