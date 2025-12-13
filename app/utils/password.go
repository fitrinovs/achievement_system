package utils

import "golang.org/x/crypto/bcrypt"

const (
	CostOfHashing = 12
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), CostOfHashing)
	return string(hash), err
}

func VerifyPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}
