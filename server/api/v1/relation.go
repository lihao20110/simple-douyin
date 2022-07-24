package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	"github.com/lihao20110/simple-douyin/server/model/system"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
	"github.com/lihao20110/simple-douyin/server/utils"
	"go.uber.org/zap"
)

type RelationApi struct {
}

//RelationAction 用户关注接口
func (r *RelationApi) RelationAction(c *gin.Context) {
	//1.获取请求参数
	toUserIDStr := c.Query("to_user_id")
	toUserID, err := strconv.ParseUint(toUserIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "to_user_id is not legal",
		})
		return
	}
	actionTypeStr := c.Query("action_type")
	if actionTypeStr != "1" && actionTypeStr != "2" { //1-关注，2-取消关注
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "action_type is not exist",
		})
		return
	}
	userID := c.GetUint64("user_id")
	//2.service层处理，返回响应
	if actionTypeStr == "1" { //1-关注，2-取消关注
		if err := relationService.AddRelationAction(userID, toUserID); err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 0,
			StatusMsg:  "add follow action success",
		})
		return
	} else { //2-取消关注
		if err := relationService.CancelRelationAction(userID, toUserID); err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 0,
			StatusMsg:  "cancel follow action success",
		})
		return
	}
}

//FollowList 关注者列表接口
func (r *RelationApi) FollowList(c *gin.Context) {
	//1.获取请求参数
	userIDStr := c.Query("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "user_id is invalid",
		})
		return
	}

	//2.service层处理
	var followUserList []system.User
	if err := relationService.GetFollowUserListRedis(userID, &followUserList); err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 402,
			StatusMsg:  err.Error(),
		})
		return
	}
	//3.返回响应
	if len(followUserList) == 0 { //关注用户为0
		global.DouYinLOG.Info("关注用户为0", zap.String("关注用户为0", "关注用户为0"))
		c.JSON(http.StatusOK, sysRes.RelationListResponse{
			Response: comRes.Response{
				StatusCode: 0,
				StatusMsg:  "关注用户为0",
			},
			UserList: []comRes.User{},
		})
		return
	}
	var (
		followResList []comRes.User
		isLogined     = false
		loginUserID   uint64
		isFollowList  []bool
		isFollow      = false
	)
	if token := c.Query("token"); token != "" { //判断传入的token是否存在，合法
		claims, err := utils.ParseToken(token)
		if err == nil && claims.ExpiresAt.UnixMilli()/1000-time.Now().UnixMilli()/1000 > 0 { //token合法,且在有效期内
			isLogined = true
			loginUserID = claims.UserID
		} else {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusInternalServerError, comRes.Response{
				StatusCode: 500,
				StatusMsg:  err.Error(),
			})
			return
		}
	}
	if isLogined { //用户处于登录状态
		userIDList := make([]uint64, len(followUserList))
		for i, user := range followUserList {
			userIDList[i] = user.ID
		}
		isFollowList, err = relationService.GetFollowStatusList(loginUserID, userIDList)
		if err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  err.Error(),
			})
			return
		}
	}
	//未登录状态，默认未关注
	for i, followUser := range followUserList {
		if isLogined {
			isFollow = isFollowList[i]
		}
		resUser := comRes.User{
			ID:            followUser.ID,
			Name:          followUser.Name,
			FollowCount:   followUser.FollowCount,
			FollowerCount: followUser.FollowerCount,
			IsFollow:      isFollow,
		}
		followResList = append(followResList, resUser)
	}
	c.JSON(http.StatusOK, sysRes.RelationListResponse{
		Response: comRes.Response{
			StatusCode: 0,
			StatusMsg:  "get follow list success",
		},
		UserList: followResList,
	})
}

//FollowerList 粉丝列表接口
func (r *RelationApi) FollowerList(c *gin.Context) {
	//1.获取请求参数
	userIDStr := c.Query("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "user_id is invalid",
		})
		return
	}

	//2.service层处理
	var followerUserList []system.User
	if err := relationService.GetFollowerUserListRedis(userID, &followerUserList); err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 402,
			StatusMsg:  err.Error(),
		})
		return
	}
	//3.返回响应
	if len(followerUserList) == 0 { //用户粉丝为0
		global.DouYinLOG.Info("用户粉丝为0", zap.String("用户粉丝为0", "用户粉丝为0"))
		c.JSON(http.StatusOK, sysRes.RelationListResponse{
			Response: comRes.Response{
				StatusCode: 0,
				StatusMsg:  "用户粉丝为0",
			},
			UserList: []comRes.User{},
		})
		return
	}
	var (
		followerResList []comRes.User
		isLogined       = false
		loginUserID     uint64
		isFollowList    []bool
		isFollow        = false
	)
	if token := c.Query("token"); token != "" { //判断传入的token是否存在，合法
		claims, err := utils.ParseToken(token)
		if err == nil && claims.ExpiresAt.UnixMilli()/1000-time.Now().UnixMilli()/1000 > 0 { //token合法,且在有效期内
			isLogined = true
			loginUserID = claims.UserID
		} else {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusInternalServerError, comRes.Response{
				StatusCode: 500,
				StatusMsg:  err.Error(),
			})
			return
		}
	}
	if isLogined { //用户处于登录状态
		userIDList := make([]uint64, len(followerUserList))
		for i, user := range followerUserList {
			userIDList[i] = user.ID
		}
		isFollowList, err = relationService.GetFollowStatusList(loginUserID, userIDList)
		if err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  err.Error(),
			})
			return
		}
	}
	//未登录状态，默认未关注
	for i, followerUser := range followerUserList {
		if isLogined {
			isFollow = isFollowList[i]
		}
		resUser := comRes.User{
			ID:            followerUser.ID,
			Name:          followerUser.Name,
			FollowCount:   followerUser.FollowCount,
			FollowerCount: followerUser.FollowerCount,
			IsFollow:      isFollow,
		}
		followerResList = append(followerResList, resUser)
	}
	c.JSON(http.StatusOK, sysRes.RelationListResponse{
		Response: comRes.Response{
			StatusCode: 0,
			StatusMsg:  "get follower list success",
		},
		UserList: followerResList,
	})
}
