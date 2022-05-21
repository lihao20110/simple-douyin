package initialize

import (
	"github.com/lihao20110/simple-douyin/server/global"
	"gorm.io/gorm"
)

// Gorm 初始化数据库并产生数据库全局变量

func Gorm() *gorm.DB {
	switch global.DouYinCONFIG.System.DbType {
	case "mysql":
		return GormMysql()
	default:
		return GormMysql()
	}
}

// RegisterTables 注册数据库表专用    gorm自动迁移试用
//实际建表用douyin.sql
//func RegisterTables(db *gorm.DB) {
//	err := db.AutoMigrate(
//		system.User{},
//		system.Video{},
//		system.Comment{},
//		system.Favorite{},
//		system.Relation{},
//	)
//	if err != nil {
//		//global.DouYinLOG.Error("register table failed", zap.Error(err))
//		fmt.Println("register table failed")
//		//os.Exit(0)
//		return
//	}
//	//global.DouYinLOG.Info("register table success")
//	fmt.Println("register table success")
//}
