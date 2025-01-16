package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// hashpassword
func HashPassword(password string) (string, error) {
	hashedPass_bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(hashedPass_bytes), err
}

// verify Password
func VerifyPassword(password, hash string) bool {
	fmt.Println(password)
	fmt.Println(hash)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false
	}
	return true // Password is valid if no error occurred during the comparison.
}
