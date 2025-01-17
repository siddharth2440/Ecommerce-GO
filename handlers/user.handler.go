package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
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
	userId, err := NUh.service.Get_My_Profile(uId)
	if err != nil {
		ctx.JSON(
			400,
			gin.H{
				"error": err.Error(),
			},
		)
		return
	}

	ctx.JSON(
		200,
		gin.H{
			"message": "User profile",
			"userId":  userId,
		},
	)
}
