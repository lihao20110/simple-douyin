package service

import (
	"sort"
	"sync"

	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CommentService struct {
}

func (c *CommentService) CommentList(videoId uint64) (*[]comRes.Comment, error) {
	var commentSysList *[]system.Comment
	err := global.DouYinDB.Model(system.Comment{}).Where("video_id = ?", videoId).Find(&commentSysList).Error
	if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		return nil, err
	}
	var authorId uint64
	err = global.DouYinDB.Model(system.Video{}).Select("author_id").Where("video_id = ?", videoId).First(&authorId).Error
	if err != nil {
		return nil, err
	}
	wg := sync.WaitGroup{}
	res := make([]comRes.Comment, 0, len(*commentSysList))
	for i := 0; i < len(*commentSysList); i++ {
		wg.Add(1)
		go func(com system.Comment) {
			defer wg.Done()
			var user system.User
			err = global.DouYinDB.Model(system.User{}).Where("user_id = ?", com.UserId).First(&user).Error
			if err != nil {
				return
			}
			//isFavorite, _ := ServiceGroupApp.FavoriteService.IsFavorite(com.UserId, com.VideoId)
			isFollow, _ := ServiceGroupApp.RelationService.IsFollow(com.UserId, authorId)
			comment := &comRes.Comment{
				ID: com.ID,
				User: comRes.User{
					ID:            user.ID,
					Name:          user.UserName,
					FollowCount:   user.FollowCount,
					FollowerCount: user.FollowerCount,
					IsFollow:      isFollow,
				},
				Content:    com.Content,
				CreateDate: com.CreatedAt.Format("01-02"),
			}
			res = append(res, *comment)
		}((*commentSysList)[i])
	}
	wg.Wait()
	sort.Slice(res, func(i, j int) bool {
		return res[i].CreateDate > res[j].CreateDate
	})
	return &res, nil
}

func (c *CommentService) CreateComment(userId, videoId uint64, content string) (*comRes.Comment, error) {
	// 开始事务
	tx := global.DouYinDB.Begin()
	newComment := system.Comment{
		UserId:  userId,
		VideoId: videoId,
		Content: content,
	}
	// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
	if err := tx.Create(&newComment).Error; err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		tx.Rollback() // 遇到错误时回滚事务
		return nil, err
	}
	//添加评论，更新视频评论数 +1
	if err := tx.Model(&system.Video{}).Where("video_id = ?", videoId).Update("comment_count", gorm.Expr("comment_count + 1")).Error; err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		tx.Rollback() // 遇到错误时回滚事务
		return nil, err
	}
	//提交事务
	tx.Commit()

	//返回当前添加的评论
	var user system.User
	if err := global.DouYinDB.Model(system.User{}).Where("user_id = ?", userId).First(&user).Error; err != nil {
		return nil, err
	}

	var authorId uint64
	if err := global.DouYinDB.Model(system.Video{}).Select("author_id").Where("video_id = ?", videoId).First(&authorId).Error; err != nil {
		return nil, err
	}

	isFollow, _ := ServiceGroupApp.RelationService.IsFollow(userId, authorId)

	return &comRes.Comment{
		ID: newComment.ID,
		User: comRes.User{
			ID:            user.ID,
			Name:          user.UserName,
			FollowCount:   user.FollowCount,
			FollowerCount: user.FollowerCount,
			IsFollow:      isFollow,
		},
		Content:    content,
		CreateDate: newComment.CreatedAt.Format("01-02"),
	}, nil
}

func (c *CommentService) DeleteComment(commentId, videoId uint) (*comRes.Response, error) {
	// 开始事务
	tx := global.DouYinDB.Begin()
	// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
	if err := tx.Where("comment_id = ?", commentId).Delete(&system.Comment{}).Error; err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		tx.Rollback() // 遇到错误时回滚事务
		return nil, err
	}
	//删除评论，更新视频评论数 -1
	if err := tx.Model(&system.Video{}).Where("video_id = ?", videoId).Update("comment_count", gorm.Expr("comment_count - 1")).Error; err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		tx.Rollback() // 遇到错误时回滚事务
		return nil, err
	}
	//提交事务
	tx.Commit()
	return &comRes.Response{
		StatusCode: 0,
		StatusMsg:  "delete comment success",
	}, nil
}
