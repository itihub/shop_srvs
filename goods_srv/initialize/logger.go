package initialize

import "go.uber.org/zap"

// 初始化zap
func InitLogger() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
}
