package config

type AliyunOSS struct {
	Endpoint        string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`
	AccessKeyId     string `mapstructure:"access-key-id" json:"access-key-id" yaml:"access-key-id"`
	AccessKeySecret string `mapstructure:"access-key-secret" json:"access-key-secret" yaml:"access-key-secret"`
	BucketName      string `mapstructure:"bucket-name" json:"bucket-name" yaml:"bucket-name"`
	BucketUrl       string `mapstructure:"bucket-url" json:"bucket-url" yaml:"bucket-url"`
	BasePath        string `mapstructure:"base-path" json:"base-path" yaml:"base-path"`
	AllowExt        string `mapstructure:"allow-ext" json:"allow-ext" yaml:"allow-ext"`
	MaxSize         int64  `mapstructure:"max-size" json:"max-size" yaml:"max-size"`
	PartSize        int64  `mapstructure:"part-size" json:"part-size" yaml:"part-size"`
	PartNum         int    `mapstructure:"part-num" json:"part-num" yaml:"part-num"`
}
