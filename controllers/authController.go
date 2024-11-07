package controllers

import (
	"UserSystem/cache"
	"UserSystem/models"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var jwtKey = []byte("secret_key")
var ctx = context.Background()

// JWT结构体
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// 注册用户
func Register(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 密码哈希
	if err := input.HashPassword(input.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hashing failed"})
		return
	}

	// 创建用户
	if result := models.DB.Create(&input); result.Error != nil {
		if strings.Contains(result.Error.Error(), "UNIQUE constraint failed") {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// 用户登录
func Login(c *gin.Context) {
	var input models.User
	var user models.User

	// 解析请求体中的 JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户是否存在
	if result := models.DB.Where("username = ?", input.Username).First(&user); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 验证密码
	if err := user.CheckPassword(input.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	loginAttemptsKey := fmt.Sprintf("login_attempts:%s", user.Username)

	// 获取用户的登录尝试次数
	attempts, _ := cache.RedisClient.Get(ctx, loginAttemptsKey).Int()
	if attempts >= 5 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many login attempts. Please try again later."})
		return
	}
	// 创建 JWT
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	cache.RedisClient.Del(ctx, loginAttemptsKey)
	// 返回包含 token 和 uid 的 JSON 响应
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"uid":   user.ID, // 返回数据库中的主键 id
	})
}

// 在登录失败时增加尝试次数
func incrementLoginAttempts(username string) {
	loginAttemptsKey := fmt.Sprintf("login_attempts:%s", username)
	cache.RedisClient.Incr(ctx, loginAttemptsKey)
	cache.RedisClient.Expire(ctx, loginAttemptsKey, time.Minute)
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// 鉴权中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		fmt.Println("Authorization Header: ", tokenString) // 打印 Authorization 头

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// 检查 Authorization 头是否包含 "Bearer " 前缀
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			// 去除 "Bearer " 前缀，只保留 JWT 部分
			tokenString = tokenString[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			fmt.Println(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}

// generateOTP 生成一个6位随机数字验证码
func generateOTP() string {
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(1000000)       // 生成0到999999的随机数
	return fmt.Sprintf("%06d", otp) // 格式化为6位数字字符串，不足位数时前面补0
}

func GenerateMFA(userID int) string {
	otp := generateOTP() // 假设有个生成 OTP 的函数
	otpKey := fmt.Sprintf("mfa:%d", userID)
	cache.RedisClient.Set(ctx, otpKey, otp, 5*time.Minute) // 设置 5 分钟有效期
	return otp
}

func ValidateMFA(userID int, userOTP string) bool {
	otpKey := fmt.Sprintf("mfa:%d", userID)
	storedOTP, err := cache.RedisClient.Get(ctx, otpKey).Result()
	if err != nil || storedOTP != userOTP {
		return false
	}
	cache.RedisClient.Del(ctx, otpKey) // 验证成功后删除 OTP
	return true
}
