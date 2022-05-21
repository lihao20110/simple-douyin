package main

import (
	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/lihao20110/simple-douyin/server/initialize"
	"go.uber.org/zap"
)

func main() {
	global.DouYinVP = initialize.Viper() //初始化viper,加载配置文件
	global.DouYinLOG = initialize.Zap()  // 初始化zap日志库
	zap.ReplaceGlobals(global.DouYinLOG)
	global.DouYinDB = initialize.Gorm() // gorm连接Mysql数据库
	if global.DouYinDB != nil {
		// 程序结束前关闭数据库链接
		db, _ := global.DouYinDB.DB()
		defer db.Close()
	} else {
		panic("数据库连接失败")
	}
	global.DouYinRedis = initialize.Redis() //连接Redis数据库
	if global.DouYinRedis == nil {
		panic("redis数据库连接失败")
	}

	r := initialize.InitRouters()
	r.Run(":8080")
}
