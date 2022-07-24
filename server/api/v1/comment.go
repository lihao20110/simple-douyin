package v1

import (
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	"github.com/lihao20110/simple-douyin/server/model/system"
	sysReq "github.com/lihao20110/simple-douyin/server/model/system/request"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
	"github.com/lihao20110/simple-douyin/server/utils"
	"go.uber.org/zap"
)

type CommentApi struct {
}

//CommentAction 评论操作接口
func (com *CommentApi) CommentAction(c *gin.Context) {
	//1.获取请求参数
	videoIDStr := c.Query("video_id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "video_id is invalid",
		})
		return
	}
	actionTypeStr := c.Query("action_type")
	if actionTypeStr != "1" && actionTypeStr != "2" {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "action_type is not exist",
		})
		return
	}
	// 获取 userID
	userID := c.GetUint64("user_id")
	//2.service层处理
	if actionTypeStr == "1" { //1-发布评论
		content := c.Query("comment_text")
		if utf8.RuneCountInString(content) >= global.MaxCommentLength ||
			utf8.RuneCountInString(content) < global.MinCommentLength {
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 400,
				StatusMsg:  "comment length not legal",
			})
			return
		}
		commentID, err := global.DouYinIDGenerator.NextID()
		if err != nil {
			//生成ID失败
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  "generate id failed",
			})
			return
		}
		commentModel := system.Comment{
			ID:      commentID,
			UserID:  userID,
			VideoID: videoID,
			Content: content,
		}
		//3.1发布评论返回响应
		if err := commentService.AddComment(&commentModel); err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  "create comment failed",
			})
			return
		}
		//获取评论用户的信息
		var userModel system.User
		if err := userService.GetUserInfoByUserIDRedis(userID, &userModel); err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  "comment user not found",
			})
			return
		}
		//获取视频作者ID
		authorID, err := feedService.GetAuthorIDByVideoID(videoID)
		if err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  "author_id not found",
			})
			return
		}
		//判断用户是否关注视频作者
		isFollow, err := relationService.GetFollowStatus(commentModel.UserID, authorID)
		if err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  "get follow status found",
			})
			return
		}
		c.JSON(http.StatusOK, sysRes.CommentActionResponse{
			Response: comRes.Response{
				StatusCode: 0,
				StatusMsg:  "comment success",
			},
			Comment: comRes.Comment{
				ID: commentModel.ID,
				User: comRes.User{
					ID:            userModel.ID,
					Name:          userModel.Name,
					FollowCount:   userModel.FollowCount,
					FollowerCount: userModel.FollowerCount,
					IsFollow:      isFollow,
				},
				Content:    content,
				CreateDate: commentModel.CreatedAt.Format("2006-01-02 15:04"),
			},
		})
		return
	} else { //2-删除评论
		commentIDStr := c.Query("comment_id")
		commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
		if err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 401,
				StatusMsg:  "comment_id is not exist",
			})
			return
		}
		//3.2删除评论返回响应
		if err := commentService.DeleteComment(commentID, videoID); err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  "delete comment failed",
			})
			return
		}
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 0,
			StatusMsg:  "delete comment success",
		})
		return
	}
}

//CommentList 评论列表接口
func (com *CommentApi) CommentList(c *gin.Context) {
	//1.获取请求参数
	var req sysReq.CommentListRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "获取请求参数失败",
		})
		return
	}
	//2.service层处理
	var commentModelList []system.Comment
	if err := commentService.GetVideoCommentListRedis(req.VideoID, &commentModelList); err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 500,
			StatusMsg:  "get comment list failed",
		})
		return
	}

	//3.返回响应
	if len(commentModelList) == 0 {
		global.DouYinLOG.Info("目前视频评论数为0", zap.Any("video", "comments"))
		c.JSON(http.StatusOK, sysRes.CommentListResponse{
			Response: comRes.Response{
				StatusCode: 0,
				StatusMsg:  "comment num is 0",
			},
			CommentList: nil,
		})
		return
	}
	var (
		commentResList []comRes.Comment
		isLogined      = false //用户是否登录
		loginUserID    uint64
		isFollowList   []bool
		isFollow       = false
	)
	userIDList := make([]uint64, len(commentModelList))
	for i, comment := range commentModelList {
		userIDList[i] = comment.UserID
	}
	var commentUserList []system.User
	if err := userService.GetUserListByIDListRedis(&commentUserList, userIDList); err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 500,
			StatusMsg:  err.Error(),
		})
		return
	}
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
		//批量获取用户对评论作者的关注状态
		var err error
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
	for i, comment := range commentModelList {
		if isLogined {
			isFollow = isFollowList[i]
		}
		resComment := comRes.Comment{
			ID: comment.ID,
			User: comRes.User{
				ID:            commentUserList[i].ID,
				Name:          commentUserList[i].Name,
				FollowCount:   commentUserList[i].FollowCount,
				FollowerCount: commentUserList[i].FollowerCount,
				IsFollow:      isFollow,
			},
			Content:    comment.Content,
			CreateDate: comment.CreatedAt.Format("2006-01-02 15:04"),
		}
		commentResList = append(commentResList, resComment)
	}
	c.JSON(http.StatusOK, sysRes.CommentListResponse{
		Response: comRes.Response{
			StatusCode: 0,
			StatusMsg:  "get comment list success",
		},
		CommentList: commentResList,
	})
}
