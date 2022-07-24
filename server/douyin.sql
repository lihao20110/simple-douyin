# DROP DATABASE IF EXISTS `sdy`;
DROP TABLE IF EXISTS `comments`;
DROP TABLE IF EXISTS `favorites`;
DROP TABLE IF EXISTS `relations`;
DROP TABLE IF EXISTS `videos`;
DROP TABLE IF EXISTS `users`;

CREATE DATABASE IF NOT EXISTS `sdy`;
USE `sdy`; # simple-douyin 数据库  (mysql8.0)

CREATE TABLE `users`
(
    `id`         bigint unsigned    NOT NULL COMMENT '用户ID',
    `created_at` datetime DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
    `name`       varchar(32) UNIQUE NOT NULL COMMENT '用户名',
    `password`   varchar(60)        NOT NULL COMMENT '用户密码',
    UNIQUE INDEX name_index (name) COMMENT '用户名索引',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT '用户表';


CREATE TABLE `videos`
(
    `id`         bigint unsigned NOT NULL COMMENT '视频ID',
    `title`      varchar(32)     NOT NULL COMMENT '视频标题',
    `created_at` datetime DEFAULT NULL COMMENT '视频发布创建时间',
    `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
    `author_id`  bigint unsigned NOT NULL COMMENT '视频发布者ID',
    `play_url`   varchar(255)    NOT NULL COMMENT '视频源地址',
    `cover_url`  varchar(255)    NOT NULL COMMENT '视频封面地址',
    INDEX videos_user_index (`author_id`) COMMENT '视频作者索引',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT '视频表';



CREATE TABLE `relations`
(
    `id`           bigint unsigned NOT NULL COMMENT '关系表ID',
    `created_at`   datetime                 DEFAULT NULL COMMENT '创建时间',
    `updated_at`   datetime                 DEFAULT NULL COMMENT '更新时间',
    `deleted_at`   datetime                 DEFAULT NULL COMMENT '删除时间',
    `from_user_id` bigint unsigned NOT NULL COMMENT '关注者ID',
    `to_user_id`   bigint unsigned NOT NULL COMMENT '被关注者ID',
    `is_follow`    tinyint(1)      NOT NULL DEFAULT '0' COMMENT '是否关注,0-未关注，1-关注',
    INDEX rel_from_to_index (from_user_id, to_user_id) COMMENT '关注者与被关注者 联合索引',
    INDEX to_index (to_user_id, is_follow) COMMENT '粉丝列表 联合索引',
    INDEX from_index (from_user_id, is_follow) COMMENT '关注列表 联合索引',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT '关系表';



CREATE TABLE `favorites`
(
    `id`          bigint unsigned NOT NULL COMMENT '点赞表ID',
    `created_at`  datetime                 DEFAULT NULL COMMENT '创建时间',
    `updated_at`  datetime                 DEFAULT NULL COMMENT '更新时间',
    `deleted_at`  datetime                 DEFAULT NULL COMMENT '删除时间',
    `user_id`     bigint unsigned NOT NULL COMMENT '点赞用户ID',
    `video_id`    bigint unsigned NOT NULL COMMENT '被点赞视频ID',
    `is_favorite` tinyint(1)      NOT NULL DEFAULT '0' COMMENT '是否点赞,0-未点赞，1-点赞',
    INDEX fav_uv_index (user_id, video_id) COMMENT '点赞用户与被点赞视频 联合索引',
    INDEX video_index (user_id, is_favorite) COMMENT '用户点赞视频列表 联合索引',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT '点赞表';



CREATE TABLE `comments`
(
    `id`         bigint unsigned NOT NULL COMMENT '评论表ID',
    `created_at` datetime     DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime     DEFAULT NULL COMMENT '更新时间',
    `deleted_at` datetime     DEFAULT NULL COMMENT '删除时间',
    `content`    varchar(255) DEFAULT NULL COMMENT '评论内容',
    `user_id`    bigint unsigned NOT NULL COMMENT '评论所属用户ID',
    `video_id`   bigint unsigned NOT NULL COMMENT '评论所属视频ID',
    INDEX com_video_index (video_id) COMMENT '评论视频ID索引',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT '评论表';
