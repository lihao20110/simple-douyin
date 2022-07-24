package service

import (
	"github.com/go-redis/redis/v8"
	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CommentService struct {
}

//GetCommentListByCommentIDListSql 当缓存未命中时，根据评论ID列表从数据库查询数据，并写入缓存
func (c *CommentService) GetCommentListByCommentIDListSql(commentList *[]system.Comment, commentIDList []uint64) error {
	*commentList = make([]system.Comment, 0, len(commentIDList))
	var getCommentList []system.Comment
	if err := global.DouYinDB.Where("id in ?", commentIDList).Find(&getCommentList).Error; err != nil {
		return err
	}
	//对查询结果建立map映射关系
	mapCommentIDToComment := make(map[uint64]system.Comment, len(getCommentList))
	for _, comment := range getCommentList {
		mapCommentIDToComment[comment.ID] = comment
	}
	for _, commentID := range commentIDList {
		tmpComment := mapCommentIDToComment[commentID]
		*commentList = append(*commentList, tmpComment)
	}
	//将评论信息列表批量写入缓存
	return c.SetCommentListRedis(*commentList)
}

//GetVideoCommentListByVideoIDSql 查询视频的评论列表
func (c *CommentService) GetVideoCommentListByVideoIDSql(videoID uint64, commentList *[]system.Comment) error {
	if err := global.DouYinDB.Where("video_id = ?", videoID).Find(&commentList).Error; err != nil {
		return err
	}
	if len(*commentList) == 0 { //点赞视频数为0，直接返回
		return nil
	}
	listZ := make([]*redis.Z, 0, len(*commentList))

	for _, comment := range *commentList {
		listZ = append(listZ, &redis.Z{
			Score:  float64(comment.UpdatedAt.UnixMilli() / 1000),
			Member: comment.ID,
		})
	}

	//将视频的评论ID列表加入缓存
	if err := c.SetVideoCommentIDListRedis(videoID, listZ...); err != nil {
		return err
	}
	return nil
}

//AddComment 添加评论,若redis添加缓存失败将回滚
func (c *CommentService) AddComment(commentModel *system.Comment) error {
	return global.DouYinDB.Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作
		if err := tx.Create(commentModel).Error; err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			return err // 遇到错误时回滚事务
		}
		if err := c.AddCommentRedis(commentModel); err != nil {
			return err
		}
		return nil
	})
}

//DeleteComment 删除评论,若操作redis缓存失败将回滚
func (c *CommentService) DeleteComment(commentID, videoID uint64) error {
	return global.DouYinDB.Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作
		if err := tx.Where("id = ?", commentID).Delete(&system.Comment{}).Error; err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			return err // 遇到错误时回滚事务
		}
		if err := c.DeleteCommentRedis(commentID, videoID); err != nil {
			return err
		}
		return nil
	})
}
