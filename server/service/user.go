package service

import (
	"errors"
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"github.com/lihao20110/simple-douyin/server/utils"
	"gorm.io/gorm"
)

type UserService struct {
}

//GetUserListByUserIDListSql 当缓存未命中时，根据用户ID列表从数据库中查询对应的用户信息，并写入缓存
func (u *UserService) GetUserListByUserIDListSql(userList *[]system.User, userIDList []uint64) error {
	*userList = make([]system.User, 0, len(userIDList))
	var getUserList []system.User
	if err := global.DouYinDB.Where("id in ?", userIDList).Find(&getUserList).Error; err != nil {
		return err
	}
	//对查询结果建立map映射关系
	mapUserIDToUser := make(map[uint64]system.User, len(getUserList))
	for i, user := range getUserList {
		var status = true
		//查询follow_count
		if err := global.DouYinDB.Model(&system.Relation{}).Where("from_user_id = ? and is_follow = ?", user.ID, &status).Count(&getUserList[i].FollowCount).Error; err != nil {
			return err
		}
		//查询follower_count
		if err := global.DouYinDB.Model(&system.Relation{}).Where("to_user_id = ? and is_follow = ?", user.ID, &status).Count(&getUserList[i].FollowerCount).Error; err != nil {
			return err
		}
		mapUserIDToUser[user.ID] = getUserList[i]
	}
	for _, userID := range userIDList {
		if v, ok := mapUserIDToUser[userID]; ok {
			*userList = append(*userList, v)
		}
	}
	//3.将用户信息批量写入缓存
	return u.SetUserListToRedis(userList)
}

//GetUserInfoByUserIDListSql 当缓存未命中时，根据用户ID从数据库中查询对应的用户信息，并写入缓存
func (u *UserService) GetUserInfoByUserIDListSql(userID uint64, userInfo *system.User) error {
	if err := global.DouYinDB.Where("id = ?", userID).First(userInfo).Error; err != nil {
		return err
	}
	var status = true
	//查询follow_count
	if err := global.DouYinDB.Model(&system.Relation{}).Where("from_user_id = ? and is_follow = ?", userID, &status).Count(&userInfo.FollowCount).Error; err != nil {
		return err
	}
	//查询follower_count
	if err := global.DouYinDB.Model(&system.Relation{}).Where("to_user_id = ? and is_follow = ?", userID, &status).Count(&userInfo.FollowerCount).Error; err != nil {
		return err
	}
	return nil
}
func (u *UserService) IsLegal(username, password string) error {
	if utf8.RuneCountInString(username) < global.MinUserLength || utf8.RuneCountInString(username) > global.MaxUserLength { //用户名长度
		return utils.ErrorUserNameNull
	}
	if ok, _ := regexp.MatchString(global.PassWordRegexp, password); !ok { //验证密码合法性
		return utils.ErrorPasswordNull
	}
	return nil
}

//Register 用户注册
func (u *UserService) Register(username, password string) (*system.User, error) {
	//用户名已注册缓存处理
	keyUsername := fmt.Sprintf(utils.UsernameRegisteredPattern, username)
	if result, err := global.DouYinRedis.Expire(global.DouYinCONTEXT, keyUsername, utils.UsernameRegisteredExpire).Result(); err == nil && result {
		return nil, utils.ErrorUserExit
	}
	//1.检查用户名是否已注册, 用户名需要保证唯一
	var user system.User
	if err := global.DouYinDB.Where("name = ? ", username).First(&user).Error; err == nil {
		//用户已存在，设置用户已注册缓存
		global.DouYinRedis.Set(global.DouYinCONTEXT, keyUsername, "1", utils.UsernameRegisteredExpire)
		return nil, utils.ErrorUserExit
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	//2.用户名不存在，注册该用户,加密密码存储数据库
	user.Name = username
	user.PassWord = utils.BcryptHash(password) //使用 bcrypt 对密码进行加密
	var err error
	user.ID, err = global.DouYinIDGenerator.NextID() //id生成
	if err != nil {
		return nil, errors.New("id generate failed")
	}
	err = global.DouYinDB.Create(&user).Error //存储到数据库
	if err != nil {
		return nil, err
	}
	return &user, nil
}

//Login 用户登录
func (u *UserService) Login(username, password string) (*system.User, error) {
	//先查缓存
	keyUserLoginEmpty := fmt.Sprintf(utils.UserLoginEmptyPattern, username)
	if result, err := global.DouYinRedis.Expire(global.DouYinCONTEXT, keyUserLoginEmpty, utils.UserLoginEmptyExpire).Result(); err == nil && result {
		return nil, errors.New("username not exists")
	}
	//1.检查用户是否存在
	var user system.User
	if err := global.DouYinDB.Where("name = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			global.DouYinRedis.Set(global.DouYinCONTEXT, keyUserLoginEmpty, "1", utils.UserLoginEmptyExpire)
			return nil, utils.ErrorUserNotExit
		}
		return nil, err
	}
	//2.用户密码验证是否正确
	if ok := utils.BcryptCheck(password, user.PassWord); !ok {
		return nil, utils.ErrorPasswordError
	}
	//3.成功，返回user
	return &user, nil
}
