package initialize

import (
	"context"
	"fmt"
	"log"
	"os"
	"shop_srvs/goods_srv/global"
	"shop_srvs/goods_srv/model"

	"github.com/olivere/elastic/v7"
)

func InitElastic() {
	// 初始化es连接
	host := fmt.Sprintf("http://%s:%d", global.ServiceConfig.ElasticInfo.Host, global.ServiceConfig.ElasticInfo.Port)
	logger := log.New(os.Stdout, "shop", log.LstdFlags)
	var err error
	global.ElasticClient, err = elastic.NewClient(elastic.SetURL(host), elastic.SetSniff(false), elastic.SetTraceLog(logger))
	if err != nil {
		panic(err)
	}

	// 新建mapping和index
	exists, err := global.ElasticClient.IndexExists(model.EsGoods{}.GetIndexName()).Do(context.Background())
	if err != nil {
		print(err)
	}
	if !exists {
		_, err = global.ElasticClient.CreateIndex(model.EsGoods{}.GetIndexName()).BodyString(model.EsGoods{}.GetMapping()).Do(context.Background())
		if err != nil {
			print(err)
		}
	}
}
