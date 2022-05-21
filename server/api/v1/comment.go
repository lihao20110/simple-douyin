package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	sysReq "github.com/lihao20110/simple-douyin/server/model/system/request"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
	"go.uber.org/zap"
)

type CommentApi struct {
}

// CommentAction no practical effect, just check if token is valid
func (com *CommentApi) CommentAction(c *gin.Context) {
	//1.获取请求参数
	videoIDStr := c.Query("video_id")
	videoId, err := strconv.ParseUint(videoIDStr, 10, 64)
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
	userId := c.GetUint64("user_id")
	//2.service层处理
	if actionTypeStr == "1" { //1-发布评论
		content := c.Query("comment_text")
		comment, err := commentService.CreateComment(userId, videoId, content)
		//3.1发布评论返回响应
		if err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  "create comment failed",
			})
			return
		}
		c.JSON(http.StatusOK, sysRes.CommentActionResponse{
			Response: comRes.Response{
				StatusCode: 0,
				StatusMsg:  "comment success",
			},
			Comment: *comment,
		})
		return
	} else { //2-删除评论
		commentIdStr := c.Query("comment_id")
		if commentIdStr == "" || commentIdStr == "0" {
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 401,
				StatusMsg:  "comment_id is not exist",
			})
			return
		}
		commentId, _ := strconv.ParseInt(commentIdStr, 10, 64)
		//3.2删除评论返回响应
		res, err := commentService.DeleteComment(uint(commentId), uint(videoId))
		if err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  "delete comment failed",
			})
			return
		}
		c.JSON(http.StatusOK, res)
		return
	}
}

// CommentList all videos have same demo comment list
func (com *CommentApi) CommentList(c *gin.Context) {
	//1.获取请求参数
	// videoIDStr := c.Query("video_id")
	// videoId, _ := strconv.ParseInt(videoIDStr, 10, 64)
	var req sysReq.CommentListRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "获取请求参数失败",
		})
		return
	}

	//2.service层处理
	commentList, err := commentService.CommentList(req.VideoID)
	//3.返回响应
	if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 500,
			StatusMsg:  "get comment list failed",
		})
		return
	}
	c.JSON(http.StatusOK, sysRes.CommentListResponse{
		Response: comRes.Response{
			StatusCode: 0,
			StatusMsg:  "get comment list success",
		},
		CommentList: *commentList,
	})
}
