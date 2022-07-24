package service

import (
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"github.com/lihao20110/simple-douyin/server/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RelationService struct {
}

//GetFollowUserListSql 查询用户的关注列表
func (r *RelationService) GetFollowUserListSql(fromUserID uint64, followUserList *[]system.User) error {
	var relationList []system.Relation
	var status = true
	if err := global.DouYinDB.Where("from_user_id = ? and is_follow = ?", fromUserID, &status).Find(&relationList).Error; err != nil {
		return err
	}
	if len(relationList) == 0 { //关注数为0，直接返回
		return nil
	}
	var (
		listZ      = make([]*redis.Z, 0, len(relationList))
		userIDList = make([]uint64, 0, len(relationList))
	)
	for _, relation := range relationList {
		userIDList = append(userIDList, relation.ToUserID)
		listZ = append(listZ, &redis.Z{
			Score:  float64(relation.UpdatedAt.UnixMilli() / 1000),
			Member: relation.ToUserID,
		})
	}
	//根据用户ID列表从缓存中查找用户信息
	if err := ServiceGroupApp.UserService.GetUserListByIDListRedis(followUserList, userIDList); err != nil {
		return err
	}
	//将用户关注的用户ID列表加入缓存
	if err := r.SetFollowUserIDListRedis(fromUserID, listZ...); err != nil {
		return err
	}
	return nil
}

//GetFollowerUserListSql 查询用户的粉丝列表
func (r *RelationService) GetFollowerUserListSql(toUserID uint64, followerUserList *[]system.User) error {
	var relationList []system.Relation
	var status = true
	if err := global.DouYinDB.Where("to_user_id = ? and is_follow = ?", toUserID, &status).Find(&relationList).Error; err != nil {
		return err
	}
	if len(relationList) == 0 { //关注数为0，直接返回
		return nil
	}
	var (
		listZ      = make([]*redis.Z, 0, len(relationList))
		userIDList = make([]uint64, 0, len(relationList))
	)
	for _, relation := range relationList {
		userIDList = append(userIDList, relation.FromUserID)
		listZ = append(listZ, &redis.Z{
			Score:  float64(relation.UpdatedAt.UnixMilli() / 1000),
			Member: relation.ToUserID,
		})
	}
	//根据用户ID列表从缓存中查找用户信息
	if err := ServiceGroupApp.UserService.GetUserListByIDListRedis(followerUserList, userIDList); err != nil {
		return err
	}
	//将用户的粉丝ID列表加入缓存
	if err := r.SetFollowerUserIDListRedis(toUserID, listZ...); err != nil {
		return err
	}
	return nil
}

//AddRelationAction 关注操作
func (r *RelationService) AddRelationAction(fromUserID, toUserID uint64) error {
	//查缓存
	followStatus, err := r.GetUserFollowStatusByUserIDRedis(fromUserID, toUserID)
	if err == nil && followStatus { //缓存中存在已关注记录
		return errors.New("重复关注")
	}
	//查找是否有一个被取消关注的记录或无记录
	var follow system.Relation
	err = global.DouYinDB.Where("from_user_id = ? and to_user_id = ?", fromUserID, toUserID).First(&follow).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil && (*follow.IsFollow) == true { //已关注过，不重复关注
		return errors.New("重复关注")
	}
	return global.DouYinDB.Transaction(func(tx *gorm.DB) error {
		if errors.Is(err, gorm.ErrRecordNotFound) { //无记录
			id, err := global.DouYinIDGenerator.NextID()
			if err != nil {
				return errors.New("id generate failed")
			}
			var status = true
			follow = system.Relation{
				ID:         id,
				FromUserID: fromUserID,
				ToUserID:   toUserID,
				IsFollow:   &status, //关注
			}
			if err := tx.Create(&follow).Error; err != nil {
				global.DouYinLOG.Error(err.Error(), zap.Error(err))
				return err
			}
		} else { //取消关注状态变为关注状态
			var status = true
			if err = tx.Model(&system.Relation{}).Select("is_follow").Where("from_user_id = ? and to_user_id = ?", fromUserID, toUserID).Update("is_follow", &status).Error; err != nil {
				global.DouYinLOG.Error(err.Error(), zap.Error(err))
				return err
			}
		}
		//删除关注列表为空值的处理
		keyEmpty := fmt.Sprintf(utils.FollowEmptyPattern, fromUserID)
		global.DouYinRedis.Del(global.DouYinCONTEXT, keyEmpty)
		//AddUserFollowRedis 关注成功后，如果缓存存在,更新用户关注数,用户关注列表
		r.AddUserFollowRedis(fromUserID, toUserID)
		return nil
	})
}

//CancelRelationAction 取消关注
func (r *RelationService) CancelRelationAction(fromUserID, toUserID uint64) error {
	//查缓存
	followStatus, err := r.GetUserFollowStatusByUserIDRedis(fromUserID, toUserID)
	if err == nil && !followStatus { //缓存中不存在已关注记录
		return errors.New("重复取消关注")
	}
	//查找是否有一个被已关注的记录
	var follow system.Relation
	err = global.DouYinDB.Where("from_user_id = ? and to_user_id = ?", fromUserID, toUserID).First(&follow).Error
	if err != nil { //执行出错或者无记录
		return err
	}
	if err == nil && (*follow.IsFollow) == false { //已取消关注过，不重复取消关注
		return errors.New("重复取消关注")
	}
	return global.DouYinDB.Transaction(func(tx *gorm.DB) error {
		//关注状态变为取消关注状态
		var status = false
		if err = tx.Model(&system.Relation{}).Select("is_follow").Where("from_user_id = ? and to_user_id = ?", fromUserID, toUserID).Update("is_follow", &status).Error; err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			return err
		}
		//SubUserFollowRedis 取消关注成功后，如果缓存存在,更新用户关注数,用户关注列表
		r.SubUserFollowRedis(fromUserID, toUserID)
		return nil
	})
}

//GetFollowStatus 获取当前用户对某用户的关注状态
func (r *RelationService) GetFollowStatus(fromUserID, toUserID uint64) (bool, error) {
	//查缓存
	followStatus, err := r.GetUserFollowStatusByUserIDRedis(fromUserID, toUserID)
	if err == nil && followStatus { //缓存中存在已关注记录
		return true, nil
	}
	if err == nil && !followStatus { //缓存中不存在已关注记录
		return false, nil
	}
	//err!=nil,缓存不存在，查数据库
	var relation system.Relation
	err = global.DouYinDB.Where("from_user_id = ? and to_user_id = ?", fromUserID, toUserID).First(&relation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err != nil {
		global.DouYinLOG.Error("check IsFollow failed", zap.Error(err))
		return false, err
	}
	return *relation.IsFollow, nil
}

//GetFollowStatusList 获取用户对一批视频作者的关注状态
func (r *RelationService) GetFollowStatusList(userID uint64, authorIDList []uint64) ([]bool, error) {
	status := make([]bool, len(authorIDList))
	var followUserIDList []uint64
	if err := r.GetFollowUserIDListRedis(userID, &followUserIDList); err != nil {
		return nil, err
	}
	if len(followUserIDList) == 0 { //点赞列表为空
		return status, nil
	}
	mapFavoriteID := make(map[uint64]struct{}, 0)
	for _, v := range followUserIDList {
		mapFavoriteID[v] = struct{}{}
	}
	for i, v := range authorIDList {
		if _, ok := mapFavoriteID[v]; ok {
			status[i] = true
		}
	}
	return status, nil
}
