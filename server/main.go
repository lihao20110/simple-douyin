package main

import (
	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/lihao20110/simple-douyin/server/initialize"
	"github.com/lihao20110/simple-douyin/server/service"
	"go.uber.org/zap"
)

func main() {
	global.DouYinIDGenerator = initialize.SonyFlake() //雪花算法分布式ID生成器
	global.DouYinVP = initialize.Viper()              //初始化viper,加载配置文件
	global.DouYinLOG = initialize.Zap()               // 初始化zap日志库
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
	//服务启动，先主动查询feed,导入缓存
	if err := service.ServiceGroupApp.FeedService.SetFeedCache(); err != nil {
		panic(err.Error())
	}
	r := initialize.InitRouters()
	if err := r.Run(":8080"); err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		return
	}
}
