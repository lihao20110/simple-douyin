package config

type System struct {
	DbType   string `mapstructure:"db-type" json:"db-type" yaml:"db-type"`       //数据库类型，如mysql
	OpenOss  bool   `mapstructure:"open-oss" json:"open-oss" yaml:"open-oss"`    //是否开始阿里云oss对象存储服务
	StaticIp string `mapstructure:"static-ip" json:"static-ip" yaml:"static-ip"` //本地存放视频和图片的url服务器地址前缀
}
