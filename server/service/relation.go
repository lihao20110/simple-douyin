package service

import (
	"errors"
	"sort"
	"sync"

	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RelationService struct {
}

// FollowList:关注列表
func (r *RelationService) FollowList(userId uint64) (*[]comRes.User, error) {
	var toUsers []uint
	err := global.DouYinDB.Model(system.Relation{}).Select("to_user_id").Where("from_user_id = ? and is_mutual = ?", userId, 1).Find(&toUsers).Error
	if err != nil {
		return nil, err
	}
	var userList *[]system.User
	err = global.DouYinDB.Model(system.User{}).Where("user_id IN (?)", toUsers).Find(&userList).Error
	if err != nil {
		return nil, err
	}
	wg := sync.WaitGroup{}
	res := make([]comRes.User, 0, len(*userList))
	for i := 0; i < len(*userList); i++ {
		wg.Add(1)
		go func(userSys system.User) {
			defer wg.Done()
			user := &comRes.User{
				ID:            userSys.ID,
				Name:          userSys.UserName,
				FollowCount:   userSys.FollowCount,
				FollowerCount: userSys.FollowerCount,
				IsFollow:      true,
			}
			res = append(res, *user)
		}((*userList)[i])
	}
	wg.Wait()
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return &res, nil
}

//FollowerList:粉丝列表
func (r *RelationService) FollowerList(userId uint64) (*[]comRes.User, error) {
	var fromUsers []uint
	err := global.DouYinDB.Model(system.Relation{}).Select("from_user_id").Where("to_user_id = ? and is_mutual = ?", userId, 1).Find(&fromUsers).Error
	if err != nil {
		return nil, err
	}
	var userList *[]system.User
	err = global.DouYinDB.Model(system.User{}).Where("user_id IN (?)", fromUsers).Find(&userList).Error
	if err != nil {
		return nil, err
	}
	wg := sync.WaitGroup{}
	res := make([]comRes.User, 0, len(*userList))
	for i := 0; i < len(*userList); i++ {
		wg.Add(1)
		go func(userSys system.User) {
			defer wg.Done()
			isFollow, _ := ServiceGroupApp.RelationService.IsFollow(userId, userSys.ID)
			user := &comRes.User{
				ID:            userSys.ID,
				Name:          userSys.UserName,
				FollowCount:   userSys.FollowCount,
				FollowerCount: userSys.FollowerCount,
				IsFollow:      isFollow,
			}
			res = append(res, *user)
		}((*userList)[i])
	}
	wg.Wait()
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return &res, nil
}

//RelationAction:关注操作
func (r *RelationService) RelationAction(fromUserId, toUserId uint64) (*comRes.Response, error) {
	// 开始事务
	tx := global.DouYinDB.Begin()
	//查找是否有一个被取消关注的记录
	var relation system.Relation
	// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
	err := tx.Where("from_user_id = ? and to_user_id = ?", fromUserId, toUserId).First(&relation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { //无记录
		relation = system.Relation{
			FromUserId: fromUserId,
			ToUserId:   toUserId,
			IsMutual:   1, //1-关注
		}
		if err := tx.Create(&relation).Error; err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			tx.Rollback() // 遇到错误时回滚事务
			return nil, err
		}
	} else if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		tx.Rollback() // 遇到错误时回滚事务
		return nil, err
	} else { //在原有记录上更新
		if err := tx.Model(&system.Relation{}).Where("from_user_id = ? and to_user_id = ?", fromUserId, toUserId).Update("is_mutual", 1).Error; err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			tx.Rollback() // 遇到错误时回滚事务
			return nil, err
		}
	}
	//关注成功后，更新关注数
	if err := tx.Model(&system.User{}).Where("user_id = ?", fromUserId).Update("follow_count", gorm.Expr("follow_count + 1")).Error; err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		tx.Rollback() // 遇到错误时回滚事务
		return nil, err
	}
	//关注成功后，更新粉丝数
	if err := tx.Model(&system.User{}).Where("user_id = ?", toUserId).Update("follower_count", gorm.Expr("follower_count + 1")).Error; err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		tx.Rollback() // 遇到错误时回滚事务
		return nil, err
	}
	//提交事务
	tx.Commit()
	return &comRes.Response{
		StatusCode: 0,
		StatusMsg:  "relation action success",
	}, nil
}

// CancelRelationAction:取消关注
func (r *RelationService) CancelRelationAction(fromUserId, toUserId uint64) (*comRes.Response, error) {
	// 开始事务
	tx := global.DouYinDB.Begin()

	// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
	//在原有记录上更新
	if err := tx.Model(&system.Relation{}).Where("from_user_id = ? and to_user_id = ?", fromUserId, toUserId).Update("is_mutual", 0).Error; err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		tx.Rollback() // 遇到错误时回滚事务
		return nil, err
	}
	//取消关注成功后，更新关注数
	if err := tx.Model(&system.User{}).Where("user_id = ?", fromUserId).Update("follow_count", gorm.Expr("follow_count - 1")).Error; err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		tx.Rollback() // 遇到错误时回滚事务
		return nil, err
	}
	//取消关注成功后，更新粉丝数
	if err := tx.Model(&system.User{}).Where("user_id = ?", toUserId).Update("follower_count", gorm.Expr("follower_count - 1")).Error; err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		tx.Rollback() // 遇到错误时回滚事务
		return nil, err
	}
	//提交事务
	tx.Commit()
	return &comRes.Response{
		StatusCode: 0,
		StatusMsg:  "cancel relation action success",
	}, nil
}

// IsFollow:判断用户是否关注了该视频作者
func (r *RelationService) IsFollow(fromUserId, toUserId uint64) (bool, error) {
	if fromUserId <= 0 || toUserId <= 0 {
		return false, nil
	}
	var relation system.Relation
	err := global.DouYinDB.Where("from_user_id = ? and to_user_id = ?", fromUserId, toUserId).First(&relation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err != nil {
		global.DouYinLOG.Error("check IsFollow failed", zap.Error(err))
		return false, err
	}
	if relation.IsMutual == 1 { //0-未关注，1-关注
		return true, nil
	}
	return false, nil
}
