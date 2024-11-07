package controllers

import (
	"UserSystem/models" // 替换为你的项目路径
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupRouterAndDatabase()

	// 添加路由
	router.POST("/register", func(c *gin.Context) {
		Register(c)
	})

	body := map[string]string{
		"username": "newuser",
		"password": "password123",
	}
	jsonValue, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// 验证数据库中是否有新用户
	var user models.User
	db.Where("username = ?", "newuser").First(&user)
	assert.Equal(t, "newuser", user.Username)
}

func TestRegisterDuplicateUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupRouterAndDatabase()

	db.Create(&models.User{Username: "duplicateuser", Password: "password123"})

	router.POST("/register", func(c *gin.Context) {
		Register(c)
	})

	body := map[string]string{
		"username": "duplicateuser",
		"password": "password123",
	}
	jsonValue, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 更新预期状态码为 409
	assert.Equal(t, http.StatusConflict, w.Code)
}
