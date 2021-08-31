package initialize

import (
	"fmt"
	"io"
	"shop_srvs/goods_srv/global"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

func InitTrace() (opentracing.Tracer, io.Closer) {
	// jaeger配置
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst, // 类型
			Param: 1,                       // 类型值
		}, // 采样配置
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true, // 发送到服务器时是否打印日志
			LocalAgentHostPort: fmt.Sprintf("%s:%d", global.ServerConfig.JaegerInfo.Host, global.ServerConfig.JaegerInfo.Port),
		}, // jaeger服务器配置
		ServiceName: global.ServerConfig.Name,
	}

	// 生成链路Tracer
	tracer, closer, err := cfg.NewTracer(jaegercfg.Logger(jaeger.StdLogger))
	if err != nil {
		panic(err)
	}
	opentracing.SetGlobalTracer(tracer)

	return tracer, closer
}
