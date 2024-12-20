package main

import (
	"UserSystem/config"
	"UserSystem/models"
	"UserSystem/routes"
	"fmt"
	"github.com/gin-contrib/cors"
	"log"
	"os"
	"time"

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
	config.InitializeRedisConfig()
	// 从环境变量中获取数据库连接字符串
	dsn := os.Getenv("DATABASE_URL")
	//front := os.Getenv("FRONT_URL")
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

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8081", "http://af3a3786a38fe426c824d75ac35c9920-2006375438.ap-southeast-1.elb.amazonaws.com:8087"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	routes.AuthRoutes(r)
	// 获取端口号，默认为8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	//port := "8082"

	// 启动服务器
	r.Run(fmt.Sprintf(":%s", port))
	//r.Run(":8080")
}
