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

// Update the UserProfile
func (NUh *User_Handler_Struct) Update_USER_Profile(ctx *gin.Context) {
	// user_chan := make(chan *domain.User, 32)
	// err_chan := make(chan error, 32)

	var to_update_user domain.To_update_user
	if err := ctx.ShouldBindJSON(&to_update_user); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error":   err.Error(),
				"success": false,
			})
		return
	}

	userId := ctx.GetString("userId")
	user, err := NUh.service.Update_My_Profile(&to_update_user, &userId)

	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Error updating user profile",
				"error":   err.Error(),
			})
		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"data":    user,
		})
}

// Delete the Userprofile
func (NUh *User_Handler_Struct) Delete_User_Profile(ctx *gin.Context) {
	userId := ctx.GetString("userId")

	deleted_user_info_chan := make(chan *domain.User, 32)
	err_chan := make(chan error, 32)

	go func() {
		user, err := NUh.service.Delete_My_Profile(userId)
		if err != nil {
			err_chan <- err
			return
		}

		deleted_user_info_chan <- user
	}()

	select {
	case user := <-deleted_user_info_chan:
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"message": "User profile deleted successfully",
				"data":    *user,
			})
	case err := <-err_chan:
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Error deleting user profile",
				"error":   err.Error(),
			})
	}
}

// Get the userProfile from given ID parameter
func (NUh *User_Handler_Struct) GET_USER_FROM_USERID(ctx *gin.Context) {
	userID := ctx.Param("userID")

	get_user_data := make(chan *domain.User, 32)
	err_data := make(chan error, 32)

	go func() {
		user, err := NUh.service.GET_USR_PROFILE(userID)
		if err != nil {
			err_data <- err
			return
		}
		get_user_data <- user
	}()

	select {
	case user := <-get_user_data:
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"message": "User profile",
				"data":    user,
			})
	case err := <-err_data:
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Error fetching user profile",
				"error":   err.Error(),
			})
	}

}

// TODO := to fix and Work this and also chk this

// Getting the n number of users
func (NUh *User_Handler_Struct) GET_RANDOM_USERS(ctx *gin.Context) {
	userNum := 1

	get_random_users := make(chan *[]domain.User, 32)
	err_chan := make(chan error, 32)
	go func() {
		users, err := NUh.service.GET_RANDOM_USERS(userNum)

		if err != nil {
			err_chan <- err
			return
		}
		get_random_users <- users
	}()

	select {
	case users := <-get_random_users:
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"message": "Random users",
				"data":    users,
			})
	case err := <-err_chan:
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Error fetching random users",
				"error":   err.Error(),
			})
	}

}

// Getting the Recently joined users
func (NUh *User_Handler_Struct) GET_RECENT_USERS(ctx *gin.Context) {
	userNum := 1
	userId := ctx.GetString("userId")
	get_recently_joined_users := make(chan *[]domain.User, 32)
	err_chan := make(chan error, 32)
	go func() {
		users, err := NUh.service.GET_RECENTLY_JOINED_USERS(userNum, userId)

		if err != nil {
			err_chan <- err
			return
		}
		get_recently_joined_users <- users
	}()

	select {
	case users := <-get_recently_joined_users:
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"message": "Random users",
				"data":    users,
			})
	case err := <-err_chan:
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Error fetching recently joined users",
				"error":   err.Error(),
			})
	}

}

// Search for user using its username email (not me)
func (NUh *User_Handler_Struct) SEARCH_FOR_USERS(ctx *gin.Context) {
	userId := ctx.GetString("userId")
	query := ctx.Query("query")
	users_chan := make(chan *[]domain.User, 32)
	err_chan := make(chan error, 32)
	go func() {
		users, err := NUh.service.Search_User(query, userId)

		if err != nil {
			err_chan <- err
			return
		}
		users_chan <- users
	}()

	select {
	case users := <-users_chan:
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"message": "User not found",
				"data":    users,
			})
	case err := <-err_chan:
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Error fetching finding users",
				"error":   err.Error(),
			})
	}

}
