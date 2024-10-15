package main

import (
	"UserSystem/models"
	"UserSystem/routes"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// init 函数会在 main 函数之前执行
func init() {
	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	// 从环境变量中获取数据库连接字符串
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		// 如果没有环境变量，使用默认的PostgreSQL连接字符串
		dsn = "host=localhost user=bruce password='12345' dbname=usersystem port=5432 sslmode=disable"
	}

	// 连接PostgreSQL数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 自动迁移数据库结构go
	db.AutoMigrate(&models.User{})
	models.DB = db

	// 初始化Gin路由
	r := gin.Default()
	routes.AuthRoutes(r)

	// 获取端口号，默认为8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 启动服务器
	r.Run(fmt.Sprintf(":%s", port))
}
