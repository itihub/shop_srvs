package initialize

import (
	"fmt"
	"shop_srvs/user_srv/global"

	"github.com/spf13/viper"
)

func getEnvInfo(env string) bool {
	viper.AutomaticEnv()
	return viper.GetBool(env)
}

// 读取配置文件
func InitConfig() {
	// 从配置文件中读取对应的配置

	debug := getEnvInfo("SHOP_DEBUG")
	configFilePrefix := "config"
	configFileName := fmt.Sprintf("user_srv/%s-prod.yaml", configFilePrefix)
	if debug {
		configFileName = fmt.Sprintf("user_srv/%s-debug.yaml", configFilePrefix)
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
