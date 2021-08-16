package initialize

import (
	"encoding/json"
	"fmt"
	"shop_srvs/goods_srv/config"
	"shop_srvs/goods_srv/global"

	"go.uber.org/zap"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/spf13/viper"
)

func getEnvInfo(env string) bool {
	viper.AutomaticEnv()
	return viper.GetBool(env)
}

func InitConfig() {
	// 读取环境变量值
	fmt.Println(getEnvInfo("SHOP_DEBUG"))
	debug := getEnvInfo("SHOP_DEBUG")

	configFilePrefix := "config"
	configFileName := fmt.Sprintf("goods_srv/%s-prod.yaml", configFilePrefix)
	if debug {
		configFileName = fmt.Sprintf("goods_srv/%s-debug.yaml", configFilePrefix)
	}
	v := viper.New()
	// 设置读取文件名称
	v.SetConfigFile(configFileName)
	// 读取
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	config := &config.Config{}
	// 配置文件映射struct
	if err := v.Unmarshal(config); err != nil {
		panic(err)
	}
	zap.S().Infof("配置信息：%v", config)
	global.NacosConfig = config.Nacos

	// 从nacos读取配置信息

	// 创建serverConfig
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      global.NacosConfig.Host,
			ContextPath: "/nacos",
			Port:        global.NacosConfig.Port,
			Scheme:      "http",
		},
	}

	// 创建clientConfig
	clientConfig := constant.ClientConfig{
		NamespaceId:         global.NacosConfig.Namespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "tmp/nacos/log",
		CacheDir:            "tmp/nacos/cache",
		RotateTime:          "1h",
		MaxAge:              3,
		LogLevel:            "debug",
	}

	// 创建动态配置客户端的另一种方式 (推荐)
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		panic(err)
	}

	// 从nacos读取配置文件
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: global.NacosConfig.Dataid,
		Group:  global.NacosConfig.Group,
	})

	if err != nil {
		panic(err)
	}

	// 想要将字符串转成json,需要去设置这个struct的tag
	err = json.Unmarshal([]byte(content), &global.ServiceConfig) // json 转换成 struct
	if err != nil {
		zap.S().Fatalf("读取nacos配置文件失败：%s", err.Error())
	}
}

// 读取配置文件
func InitConfig2() {
	// 从配置文件中读取对应的配置

	debug := getEnvInfo("SHOP_DEBUG")
	configFilePrefix := "config"
	configFileName := fmt.Sprintf("goods_srv/%s-prod.yaml", configFilePrefix)
	if debug {
		configFileName = fmt.Sprintf("goods_srv/%s-debug.yaml", configFilePrefix)
	}

	v := viper.New()
	v.SetConfigFile(configFileName)
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := v.Unmarshal(&global.ServiceConfig); err != nil {
		panic(err)
	}
}
