

# Nacos  
```shell
docker run --name=nacos \
    -e MODE=standalone -e PREFER_HOST_MODE=hostname \
    -e SPRING_DATASOURCE_PLATFORM=mysql -e MYSQL_SERVICE_HOST=mysql \
    -e MYSQL_SERVICE_PORT=3306 -e MYSQL_SERVICE_DB_NAME=nacos_config \
    -e MYSQL_SERVICE_USER=root -e MYSQL_SERVICE_PASSWORD=123456 \
    -e JVM_XMS=512m -e JVM_XMX=512m -e JVM_XMN=256m  \
    --link some-mysql:mysql -p 8848:8848 -d nacos/nacos-server:2.0.2
```

# Consul  
```shell
docker run -d --name=consul \
    -p 8500:8500 -p 8300:8300 -p 8301:8301 \
    -p 8302:8302 -p 8600:8600/udp consul consul agent -dev -client=0.0.0.0 
```