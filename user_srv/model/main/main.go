package main

import (
	"crypto/md5"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"io"
	"log"
	"os"
	"shop_srvs/user_srv/model"
	"strings"
	"time"
)

// 用于生成结构
func main() {
	// 使用gorm连接到数据库

	// 设置全局的的logger, 作用：执行每个sql语句的时候会打印每一行sql
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // 慢sql阈值
			LogLevel:                  logger.Info, // Log级别
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // 禁用彩色打印
		},
	)

	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	dsn := "root:123456@tcp(local.docker.node1.com:3306)/micro_user_srv?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger, // logger全局模式
	})
	if err != nil {
		panic(err)
	}

	// 迁移 schema
	db.AutoMigrate(&model.User{})

	// 批量创建测试数据
	//options := &password.Options{16, 100, 32, sha512.New}
	//salt, encodedPwd := password.Encode("admin123", options)
	//newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)
	//for i := 0; i < 10; i++ {
	//	user := model.User{
	//		NickName: fmt.Sprintf("jimmy%d", i),
	//		Mobile: fmt.Sprintf("1560000000%d", i),
	//		Password: newPassword,
	//	}
	//	db.Save(&user)
	//}

	//fmt.Println(genMd5("123456"))
	//encodeTest()
}

// 生成MD5 避免彩虹表爆破采用加盐值来提高安全性
func genMd5(code string) string {
	// 生成MD5方式一
	//md5.Sum([]byte(code))

	// 生成MD5方式二
	Md5 := md5.New()
	_, _ = io.WriteString(Md5, code)
	return hex.EncodeToString(Md5.Sum(nil))
}

// 加盐MD5测试 使用的是第三方的github.com/anaskhan96/go-password-encoder
func encodeTest() {
	// Using the default options
	salt, encodedPwd := password.Encode("generic password", nil)
	check := password.Verify("generic password", salt, encodedPwd, nil)
	fmt.Println("salt:", salt)
	fmt.Println("MD5:", encodedPwd)
	fmt.Println(check) // true

	// Using custom options        盐值长度   生成盐值迭代次数    key长度            加密算法
	options := &password.Options{16, 100, 32, sha512.New}
	salt, encodedPwd = password.Encode("generic password", options)
	check = password.Verify("generic password", salt, encodedPwd, options)
	fmt.Println("salt:", salt)
	fmt.Println("MD5:", encodedPwd)
	fmt.Println(check) // true

	// 生成用户密码并携带加密信息以及盐值 格式：加密算法 + 盐值 + MD5
	newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)
	fmt.Println(newPassword)
	fmt.Println(len(newPassword)) // 长度不超过100

	// 校验用户密码
	passwordInfo := strings.Split(newPassword, "$") // 解析密码信息
	fmt.Println(passwordInfo)
	check = password.Verify("generic password", passwordInfo[2], passwordInfo[3], options) // 校验
	fmt.Println(check)
}
