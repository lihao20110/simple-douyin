package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
	"go.uber.org/zap"
)

type RelationApi struct {
}

func (r *RelationApi) RelationAction(c *gin.Context) {
	//1.获取请求参数
	toUserIDStr := c.Query("to_user_id")
	if toUserIDStr == "" || toUserIDStr == "0" {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "toUserIDStr is not exist",
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
	userId, _ := c.Get("user_id")
	toUserId, _ := strconv.ParseUint(toUserIDStr, 10, 64)
	userIdUint := userId.(uint64)
	//2.service层处理，返回响应
	if actionTypeStr == "1" { //1-关注，2-取消关注
		res, err := relationService.RelationAction(userIdUint, toUserId)
		if err != nil {
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  "relation action failed",
			})
			return
		}
		c.JSON(http.StatusOK, res)
		return
	} else { //2-取消关注
		res, err := relationService.CancelRelationAction(userIdUint, toUserId)
		if err != nil {
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  "cancel relation action failed",
			})
			return
		}
		c.JSON(http.StatusOK, res)
		return
	}
}

//关注者列表
func (r *RelationApi) FollowList(c *gin.Context) {
	//1.获取请求参数
	userIdStr := c.Query("user_id")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "user_id is invalid",
		})
		return
	}
	//2.service层处理
	followUserList, err := relationService.FollowList(userId)
	if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 402,
			StatusMsg:  err.Error(),
		})
		return
	}
	//3.返回响应
	if len(*followUserList) == 0 { //关注用户为0
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
	c.JSON(http.StatusOK, sysRes.RelationListResponse{
		Response: comRes.Response{
			StatusCode: 0,
			StatusMsg:  "get follow list success",
		},
		UserList: *followUserList,
	})
}

//粉丝列表
func (r *RelationApi) FollowerList(c *gin.Context) {
	//1.获取请求参数
	userIdStr := c.Query("user_id")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "user_id is invalid",
		})
		return
	}

	//2.service层处理
	followerUserList, err := relationService.FollowerList(userId)
	if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 402,
			StatusMsg:  err.Error(),
		})
		return
	}
	//3.返回响应
	if len(*followerUserList) == 0 { //粉丝用户为0
		global.DouYinLOG.Info("粉丝用户为0", zap.String("粉丝用户为0", "粉丝用户为0"))
		c.JSON(http.StatusOK, sysRes.RelationListResponse{
			Response: comRes.Response{
				StatusCode: 0,
				StatusMsg:  "粉丝用户为0",
			},
			UserList: []comRes.User{},
		})
		return
	}
	c.JSON(http.StatusOK, sysRes.RelationListResponse{
		Response: comRes.Response{
			StatusCode: 0,
			StatusMsg:  "get follower list success",
		},
		UserList: *followerUserList,
	})
}
