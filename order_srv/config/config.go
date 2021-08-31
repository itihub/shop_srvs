package config

type SrvConfig struct {
	Name string `mapstructure:"name" json:"name"`
}

type MysqlConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Name     string `mapstructure:"db" json:"db"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
}

type ConsulConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type RedisConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type RocketMQConfig struct {
	Host              string `mapstructure:"host" json:"host"`
	Port              int    `mapstructure:"port" json:"port"`
	OrderRebackTopic  string `mapstructure:"order_reback_topic" json:"order_reback_topic"`
	OrderTimeoutTopic string `mapstructure:"order_timeout_topic" json:"order_timeout_topic"`
}

type JaegerConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type ServerConfig struct {
	Host         string         `mapstructure:"host" json:"host"`
	Port         int            `mapstructure:"port" json:"port"`
	Name         string         `mapstructure:"name" json:"name"`
	Tags         []string       `mapstructure:"tags" json:"tags"`
	MysqlInfo    MysqlConfig    `mapstructure:"mysql" json:"mysql"`
	ConsulInfo   ConsulConfig   `mapstructure:"consul" json:"consul"`
	RedisInfo    RedisConfig    `mapstructure:"redis" json:"redis"`
	RocketMQInfo RocketMQConfig `mapstructure:"rocket-mq" json:"rocket-mq"`
	// 商品微服务的配置
	GoodsSrvInfo SrvConfig `mapstructure:"goods_srv" json:"goods_srv"`
	// 库存微服务的配置
	InventorySrvInfo SrvConfig    `mapstructure:"inventory_srv" json:"inventory_srv"`
	JaegerInfo       JaegerConfig `mapstructure:"jaeger" json:"jaeger"`
}

type NacosConfig struct {
	Host      string `mapstructure:"host" json:"host"`
	Port      uint64 `mapstructure:"port" json:"port"`
	Namespace string `mapstructure:"namespace" json:"namespace"`
	User      string `mapstructure:"user" json:"user"`
	Password  string `mapstructure:"password" json:"password"`
	Dataid    string `mapstructure:"dataid" json:"dataid"`
	Group     string `mapstructure:"group" json:"group"`
}

type Config struct {
	Nacos NacosConfig `mapstructure:"nacos"`
}
