package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/ecommerce/config"
)

// create a token
func Create_JWT_Token(userId string, username string, isAdmin bool) (string, error) {
	config, _ := config.SetConfig()
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id":       userId,
			"username": username,
			"isAdmin":  isAdmin,
			"exp":      time.Now().Add(time.Hour * 72).Unix(),
		},
	)

	tokenString, err := token.SignedString([]byte(config.JWT_SECRET))
	if err != nil {
		return "", err
	}

	// fmt.Println("My JWT Token String created")
	// fmt.Printf("JWT Token %s\n", tokenString)

	return tokenString, nil
}

// token verification
func JWT_Verification(tokenString string) (*jwt.Token, error) {
	config, _ := config.SetConfig()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWT_SECRET), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return token, nil
}
