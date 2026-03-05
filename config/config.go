package config

type Config struct {
	Name            string      `mapstructure:"appName"`         //服务名称
	Port            int         `mapstructure:"port"`            //运行端口
	Mysql           Mysql       `mapstructure:"mysql"`           //mysql 配置
	Log             Log         `mapstructure:"log"`             //日志配置
	UserMiniProgram MiniProgram `mapstructure:"userMiniProgram"` //用户小程序配置
	ClubMiniProgram MiniProgram `mapstructure:"clubMiniProgram"` //俱乐部小程序配置
	Mch             Mch         `mapstructure:"mch"`             //微信商户配置
	Oss             Oss         `mapstructure:"oss"`             //阿里云配置
	JWT             JWT         `mapstructure:"jwt"`             //jwt配置
	Email           Email       `mapstructure:"email"`           //邮箱配置
}

type Mysql struct {
	Host     string `mapstructure:"host"` //数据库连接字符串
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type Log struct {
	Path  string `mapstructure:"path"`
	Level int    `mapstructure:"level"`
}

type MiniProgram struct {
	AppId  string `mapstructure:"appId"`
	Secret string `mapstructure:"secret"`
}

type Mch struct {
	MchId                  string `mapstructure:"mchId"`
	ApiKey                 string `mapstructure:"apiKey"`
	SerialNumber           string `mapstructure:"serialNumber"`
	PublicKeyId            string `mapstructure:"publicKeyId"`
	OrderPayNotifyUrl      string `mapstructure:"orderPayNotifyUrl"`      //订单支付成功回调
	OrderRefundNotifyUrl   string `mapstructure:"orderRefundNotifyUrl"`   //订单退款成功回调
	DepositPayNotifyUrl    string `mapstructure:"depositPayNotifyUrl"`    //保证金充值成功回调
	TransferBillsNotifyUrl string `mapstructure:"transferBillsNotifyUrl"` //商家转账成功回调
}

type Oss struct {
	AccessKeyId     string `mapstructure:"accessKeyId"`
	AccessKeySecret string `mapstructure:"accessKeySecret"`
	EndPoint        string `mapstructure:"endpoint"`
	Bucket          string `mapstructure:"bucket"` //存储桶
}

type JWT struct {
	SigningKey string `mapstructure:"key"`
}

type Email struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	UserName string `mapstructure:"userName"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}
