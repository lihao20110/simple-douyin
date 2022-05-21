package v1

import (
	"fmt"
	"testing"

	"github.com/lihao20110/simple-douyin/server/global"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGetVideoList(t *testing.T) {
	dsn := "root:haohao@tcp(127.0.0.1:13306)/sdy?charset=utf8mb4&parseTime=True&loc=Local"
	global.DouYinDB, _ = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	videoList, err := publishService.GetVideoList(1)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(videoList)
	userInfo, err := userService.GetUserInfoById(1)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(userInfo)
}
