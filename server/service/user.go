package service

import (
	"errors"
	"regexp"
	"unicode/utf8"

	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"github.com/lihao20110/simple-douyin/server/utils"
	"gorm.io/gorm"
)

type UserService struct {
}

func (u *UserService) IsIDAndTokenMatch(id int64, token string) bool {
	userId := utils.GetUserId(token)
	if userId == 0 { //默认值为0，主键ID为0，则说明用户不存在，未登录
		return false
	}
	if id != int64(userId) {
		return false
	}
	return true
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

func (u *UserService) Register(username, password string) (*system.User, error) {
	//1.检查用户名是否已注册, 用户名需要保证唯一
	var user system.User
	if !errors.Is(global.DouYinDB.Where("user_name = ? ", username).First(&user).Error, gorm.ErrRecordNotFound) {
		return nil, utils.ErrorUserExit
	}
	//2.用户名不存在，注册该用户,加密密码存储数据库
	user.UserName = username
	user.PassWord = utils.BcryptHash(password) //使用 bcrypt 对密码进行加密
	err := global.DouYinDB.Create(&user).Error //存储到数据库
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *UserService) Login(username, password string) (*system.User, error) {
	//1.检查用户是否存在
	var user system.User
	err := global.DouYinDB.Where("user_name = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrorUserNotExit
	}
	//2.用户密码验证是否正确
	if ok := utils.BcryptCheck(password, user.PassWord); !ok {
		return nil, utils.ErrorPasswordError
	}
	//3.成功，返回user
	return &user, nil
}

func (u *UserService) GetUserInfoById(userId uint64) (*system.User, error) {
	var user system.User
	err := global.DouYinDB.Where("user_id = ?", userId).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrorUserNotExit
	}
	return &user, nil
}
