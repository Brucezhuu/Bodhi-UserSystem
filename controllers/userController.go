package controllers

import (
	"UserSystem/cache"
	"UserSystem/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetUser(c *gin.Context) {
	id := c.Param("id")
	cacheKey := fmt.Sprintf("user:%s", id)

	// 尝试从 Redis 获取用户数据
	if cachedData, err := cache.GetCache(cacheKey); err == nil {
		var user models.User
		if err := json.Unmarshal([]byte(cachedData), &user); err == nil {
			c.JSON(http.StatusOK, user)
			return
		}
	}

	// 若缓存未命中，则从数据库查询用户数据
	var user models.User
	if result := models.DB.First(&user, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 将查询到的用户数据缓存到 Redis
	userData, _ := json.Marshal(user)
	cache.SetCache(cacheKey, userData, 10*time.Minute) // 设置缓存过期时间为 10 分钟

	c.JSON(http.StatusOK, user)
}
