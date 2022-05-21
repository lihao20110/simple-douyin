package utils

import (
	"fmt"
	"testing"
)

func TestBcrypt(t *testing.T) {
	password := "123456"
	hash := BcryptHash(password)
	fmt.Println(hash)
	if ok := BcryptCheck(password, hash); ok {
		fmt.Println("success")
	}
}
