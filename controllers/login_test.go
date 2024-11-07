package controllers

import (
	"UserSystem/models" // 替换为你的项目路径
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 初始化 Gin 和数据库
func setupRouterAndDatabase() (*gin.Engine, *gorm.DB) {
	// 创建 SQLite 数据库，用于测试（内存数据库，不会持久化）
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// 自动迁移 User 表
	db.AutoMigrate(&models.User{})

	// 初始化 Gin 引擎
	r := gin.Default()
	r.POST("/login", func(c *gin.Context) {
		Login(c)
	})
	models.DB = db

	return r, db
}

// 测试用户登录成功
func TestLoginSuccess(t *testing.T) {
	// 设置 Gin 测试模式
	gin.SetMode(gin.TestMode)

	// 初始化路由和数据库
	router, db := setupRouterAndDatabase()

	// 创建测试用户
	password := "password123"
	hashedPassword, _ := models.HashPassword(password)
	user := models.User{
		Username: "testuser",
		Password: hashedPassword,
	}
	db.Create(&user)

	// 模拟请求体
	body := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	jsonValue, _ := json.Marshal(body)

	// 创建 HTTP 请求
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	// 记录响应
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 解析响应
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// 检查响应状态码和内容
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, response["token"], "Token should be returned")
	assert.Equal(t, float64(user.ID), response["uid"].(float64), "UID should match the user ID")
}

// 测试用户登录失败
func TestLoginFailure(t *testing.T) {
	// 设置 Gin 测试模式
	gin.SetMode(gin.TestMode)

	// 初始化路由和数据库
	router, _ := setupRouterAndDatabase()

	// 模拟无效的请求体
	body := map[string]string{
		"username": "invaliduser",
		"password": "wrongpassword",
	}
	jsonValue, _ := json.Marshal(body)

	// 创建 HTTP 请求
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	// 记录响应
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 解析响应
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// 检查响应状态码和内容
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "Invalid credentials", response["error"], "Error message should be 'Invalid credentials'")
}
