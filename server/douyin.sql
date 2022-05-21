USE `sdy`;    #simple-douyin 数据库

# 不用外键,多用索引

DROP TABLE IF EXISTS `comments`;
DROP TABLE IF EXISTS `favorites`;
DROP TABLE IF EXISTS `relations`;
DROP TABLE IF EXISTS `videos`;
DROP TABLE IF EXISTS `users`;

CREATE TABLE `users`(
    `user_id`   int unsigned  NOT NULL	AUTO_INCREMENT  COMMENT '用户ID',
	  `created_at` datetime     DEFAULT NULL              COMMENT '创建时间', 
		`updated_at` datetime     DEFAULT NULL              COMMENT '更新时间', 
    `deleted_at` datetime     DEFAULT NULL              COMMENT '删除时间', 
    `user_name`       varchar(32)  UNIQUE NOT NULL      COMMENT '用户名',
    `user_password`   varchar(60)     NOT NULL          COMMENT '用户密码',
    `follow_count`	int        NOT NULL DEFAULT 0       COMMENT '关注数量',
    `follower_count` int      NOT NULL DEFAULT 0        COMMENT '粉丝数量',
    UNIQUE INDEX name_index (user_name)                 COMMENT '用户名索引',
		PRIMARY KEY(`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COMMENT '用户表';



CREATE TABLE `videos`(
	  `video_id`         int unsigned    NOT NULL	AUTO_INCREMENT  COMMENT '视频ID',
		`video_title`      varchar(32)  NOT NULL                    COMMENT '视频标题',
	  `created_at` datetime     DEFAULT NULL              COMMENT '视频发布创建时间', 
    `updated_at` datetime     DEFAULT NULL              COMMENT '更新时间', 
    `deleted_at` datetime     DEFAULT NULL              COMMENT '删除时间', 
    `author_id`  int unsigned       NOT NULL            COMMENT '视频发布者ID',
    `play_url`   varchar(255) NOT NULL                  COMMENT '视频源地址',
    `cover_url`  varchar(255) NOT NULL                  COMMENT '视频封面地址',
		`favorite_count` int      NOT NULL DEFAULT 0        COMMENT '视频点赞数',       
    `comment_count`  int      NOT NULL DEFAULT 0        COMMENT '视频评论数',
    INDEX videos_create_time_index (`created_at`)       COMMENT '视频发布时间索引',
		INDEX videos_user_index (`author_id`)               COMMENT '视频作者索引',
		PRIMARY KEY(`video_id`)
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COMMENT '视频表';



CREATE TABLE `relations`(
		`relation_id`   int  unsigned   NOT NULL	AUTO_INCREMENT  COMMENT '关系表ID',
		`created_at` datetime     DEFAULT NULL       COMMENT '创建时间', 
    `updated_at` datetime     DEFAULT NULL       COMMENT '更新时间', 
    `deleted_at` datetime     DEFAULT NULL       COMMENT '删除时间', 
    `from_user_id` int unsigned   NOT NULL       COMMENT '关注者ID',
		`to_user_id`   int unsigned   NOT NULL       COMMENT '被关注者ID',
		`is_mutual`    tinyint(1)    NOT NULL DEFAULT '0'     COMMENT '是否关注,0-未关注，1-关注',
		INDEX rel_from_to_index (from_user_id,to_user_id)     COMMENT '关注者与被关注者 联合索引',
		INDEX to_index(to_user_id) COMMENT '粉丝列表索引',
    PRIMARY KEY (`relation_id`)  
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COMMENT '关系表';



CREATE TABLE `favorites`(
		`favorite_id`   int  unsigned   NOT NULL	AUTO_INCREMENT  COMMENT '点赞表ID',
		`created_at` datetime     DEFAULT NULL       COMMENT '创建时间', 
    `updated_at` datetime     DEFAULT NULL       COMMENT '更新时间', 
    `deleted_at` datetime     DEFAULT NULL       COMMENT '删除时间', 
    `user_id`  int unsigned   NOT NULL           COMMENT '点赞用户ID',
    `video_id` int unsigned  NOT NULL            COMMENT '被点赞视频ID',
		`status`   tinyint(1)    NOT NULL DEFAULT '0'     COMMENT '是否点赞,0-未点赞，1-点赞',
		INDEX fav_uv_index (user_id,video_id)        COMMENT '点赞用户与被点赞视频 联合索引',
		PRIMARY KEY (`favorite_id`)
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COMMENT '点赞表';



CREATE TABLE `comments`(
		`comment_id`   int  unsigned   NOT NULL	AUTO_INCREMENT  COMMENT '评论表ID',
		`created_at` datetime     DEFAULT NULL       COMMENT '创建时间', 
    `updated_at` datetime     DEFAULT NULL       COMMENT '更新时间', 
    `deleted_at` datetime     DEFAULT NULL       COMMENT '删除时间', 
		`content` varchar(255)    DEFAULT NULL       COMMENT '评论内容', 
    `user_id`  int unsigned   NOT NULL           COMMENT '评论所属用户ID',
    `video_id` int unsigned   NOT NULL           COMMENT '评论所属视频ID',
		INDEX com_video_index (video_id)             COMMENT '评论视频ID索引',
		INDEX videos_create_time_index (`created_at`)       COMMENT '评论发布时间索引',
		PRIMARY KEY (`comment_id`)
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COMMENT '评论表';
