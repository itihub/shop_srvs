# 商城微服务  

## Table of Contents  

- [Background](#background)  
- [Install](#install)  

## Background  

## Install    

 Nacos  
```shell
docker run --name=nacos \
    -e MODE=standalone -e PREFER_HOST_MODE=hostname \
    -e SPRING_DATASOURCE_PLATFORM=mysql -e MYSQL_SERVICE_HOST=mysql \
    -e MYSQL_SERVICE_PORT=3306 -e MYSQL_SERVICE_DB_NAME=nacos_config \
    -e MYSQL_SERVICE_USER=root -e MYSQL_SERVICE_PASSWORD=123456 \
    -e JVM_XMS=512m -e JVM_XMX=512m -e JVM_XMN=256m  \
    --link some-mysql:mysql -p 8848:8848 -d nacos/nacos-server:2.0.2
```

  Consul    
```shell
docker run -d --name=consul \
    -p 8500:8500 -p 8300:8300 -p 8301:8301 \
    -p 8302:8302 -p 8600:8600/udp consul consul agent -dev -client=0.0.0.0 
```

  Elasticsearch
```shell
# 新建es的config配置文件夹
mkdir -p /data/elasticsearch/config
# 新建es的data目录
mkdir -p /data/elasticsearch/data
# 给目录设置权限
chmod 777 -R /data/elasticsearch
# 写入配置到elasticsearch.yml中
echo "http.host: 0.0.0.0" >> /data/elasticsearch/config/elasticsearch.yml
# 安装es
docker run --name some-elasticsearch -p 9200:9200 -p 9300:9300 \
    -e "discovery.type=single-node" \
    -e ES_JAVA_OPTS="-Xms128m -Xmx256m" \
    -v /data/elasticsearch/config/elasticsearch.yml:/usr/share/elasticsearch/config/elasticsearch.yml \
    -v /data/elasticsearch/data:/usr/share/elasticsearch/data \
    -v /data/elasticsearch/plugins:/usr/share/elasticsearch/plugins \
    -d elasticsearch:7.10.1
```

  Kibana
```shell
docker run --name some-kibana \
    --link=some-elasticsearch:elasticsearch \
    -p 5601:5601 -d kibana:7.10.1
```

  Elasticsearch-analysis-ik  
```shell
#下载对应elasticsearch版本的ik分词器  
wget https://github.com/medcl/elasticsearch-analysis-ik/releases/download/v7.10.1/elasticsearch-analysis-ik-7.10.1.zip
#解压到plugins目录下并命名文件夹为ik
unzip -o -d ik elasticsearch-analysis-ik-7.10.1.zip
# 设置权限
cd /data/elasticsearch/plugins
chmod 777 -R ik
# 重启Elasticsearch容器
docker restart some-elasticsearch
```

  RocketMQ  
```shell
# 将deployment/rocketmq目录上传服务器
docker-compose up
```
