package utils

import (
	"fmt"
	"testing"
)

func TestGetCoverImage(t *testing.T) {
	coverImageName, _ := GetCoverImage("../public/boat.mp4", "../public/boat", 24)
	fmt.Println("--------------------------")
	fmt.Println(coverImageName)
}
