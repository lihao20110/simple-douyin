package initialize

import (
	"math/rand"
	"time"

	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/sony/sonyflake"
)

func SonyFlake() *sonyflake.Sonyflake {
	rand.Seed(time.Now().Unix())
	startTime, _ := time.Parse("2006-01-02 15:04:05", global.StartTime)
	sf := sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: startTime,
	})
	return sf
}
