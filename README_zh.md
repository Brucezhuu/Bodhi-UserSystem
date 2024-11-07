# Bodhi Usersystem Database Schema

## 1. `users` 表：扩展用户信息
`users` 表存储用户的基本信息，增加了更多字段以支持更丰富的用户管理功能。

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    email TEXT UNIQUE,
    phone TEXT UNIQUE,
    account_status TEXT DEFAULT 'active',          -- 用户账户状态
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP                           -- 用于软删除
);
```
- `username` 和 `password`：用于身份验证。
- `email` 和 `phone`：支持多种联系信息。
- `account_status`：表示账户状态（如活跃、暂停、禁用），便于账户管理。
- `deleted_at`：用于软删除，便于恢复数据。
## 2. `roles` 和 `user_roles` 表：支持用户角色和权限管理
`roles` 表定义了系统中的角色，`user_roles` 用于将用户与角色关联，支持多对多关系，实现基本的基于角色的访问控制（RBAC）。

```sql
CREATE TABLE roles (
                       id INTEGER PRIMARY KEY AUTOINCREMENT,
                       role_name TEXT NOT NULL UNIQUE,
                       description TEXT
);

CREATE TABLE user_roles (
                            user_id INTEGER NOT NULL,
                            role_id INTEGER NOT NULL,
                            assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                            PRIMARY KEY (user_id, role_id),
                            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
                            FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

```
- `roles` 表定义系统角色。
- `user_roles` 表实现用户和角色的多对多关联，用于分配用户权限。

## 3. `user_activity_logs` 表：用户活动日志
记录用户的登录、登出、密码更改等活动，便于审计和追踪，提升系统的安全性。
```sql
CREATE TABLE user_activity_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    activity_type TEXT NOT NULL,                   -- 活动类型，如登录、登出等
    activity_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,                               -- 记录用户 IP 地址
    user_agent TEXT,                               -- 用户设备的信息
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

```
- `activity_type`：记录活动类型（如登录、登出）。
- `ip_address` 和 `user_agent`：提供用户设备和访问信息，用于审计。

## 4. `user_mfa 表`：多因素认证支持
提供多因素认证（MFA）支持，将 OTP 等临时验证码保存在 Redis 中，并在数据库中记录配置，提升用户账户的安全性。

```sql
CREATE TABLE user_mfa (
    user_id INTEGER PRIMARY KEY,
    mfa_type TEXT NOT NULL,                        -- MFA 类型，如 SMS, email, TOTP
    mfa_secret TEXT NOT NULL,                      -- MFA 密钥或验证码
    is_enabled BOOLEAN DEFAULT FALSE,              -- MFA 是否启用
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```
- `mfa_type`：支持多种 MFA 类型。
- `mfa_secret`：存储密钥或验证码。
- `is_enabled`：表示 MFA 是否启用。

## 5. `user_preferences` 表：用户偏好设置
支持用户自定义界面和通知设置，提升用户体验。

```sql
CREATE TABLE user_preferences (
    user_id INTEGER PRIMARY KEY,
    theme TEXT DEFAULT 'light',                    -- 界面主题
    language TEXT DEFAULT 'en',                    -- 语言偏好
    receive_notifications BOOLEAN DEFAULT TRUE,    -- 通知偏好
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

```

- `theme` 和 `language`：支持自定义界面和语言设置。
- `receive_notifications`：允许用户管理通知偏好。

---

## 引入 Redis 后的缓存和数据库架构
Redis 的引入增加了缓存层，提高了系统性能、响应速度和扩展性，特别是在用户管理、访问控制和限流等方面带来了显著的优势。

### Redis 应用场景

#### 1. 用户会话管理
将用户的会话信息（如 JWT）存储在 Redis 中，方便高效地验证用户身份，提升用户体验和系统性能。

#### 2. 缓存用户信息
将频繁访问的用户数据缓存到 Redis 中，减少数据库查询，降低延迟。例如在 `GetUser` 操作中，先尝试从 Redis 获取用户信息，若缓存未命中再访问数据库并将结果存入缓存。

```go
func GetUser(c *gin.Context) {
    userID := c.Param("id")
    cacheKey := fmt.Sprintf("user:%s", userID)
    if cachedData, err := cache.GetCache(cacheKey); err == nil {
        var user models.User
        json.Unmarshal([]byte(cachedData), &user)
        c.JSON(http.StatusOK, user)
        return
    }
    // 数据库查询逻辑省略
}
```
#### 3. 限制频繁操作（限流）
使用 Redis 的计数器功能限制用户操作频率。例如在登录操作中，设置 1 分钟内最多允许 5 次登录尝试，超出限制则阻止登录请求。此方式防止暴力破解和接口滥用。
```go
func Login(c *gin.Context) {
    username := c.PostForm("username")
    loginAttemptsKey := fmt.Sprintf("login_attempts:%s", username)
    attempts, _ := cache.RedisClient.Get(ctx, loginAttemptsKey).Int()
    if attempts >= 5 {
        c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many login attempts."})
        return
    }
    cache.RedisClient.Incr(ctx, loginAttemptsKey)
    cache.RedisClient.Expire(ctx, loginAttemptsKey, time.Minute)
    // 登录逻辑省略
}

```
#### 4. 多因素认证（MFA）临时数据存储
使用 Redis 存储 OTP 等临时验证码，支持多因素认证。设置短期过期时间（TTL）确保验证码安全可靠。
```go
func GenerateMFA(userID int) string {
    otp := generateOTP() // 生成一次性密码
    otpKey := fmt.Sprintf("mfa:%d", userID)
    cache.RedisClient.Set(ctx, otpKey, otp, 5*time.Minute) // 5 分钟有效
    return otp
}
```
#### 5. 用户活动日志缓存
对于频繁访问的活动记录，使用 Redis 缓存近期的活动日志，快速查询最近的用户活动，减少数据库压力。

### Redis 带来的优势
- 性能提升：Redis 作为缓存层可以大幅减少数据库的访问次数，提高响应速度。
- 扩展性：缓存用户信息、MFA 及限流机制，使系统能够支持更高的并发和更复杂的业务需求。
- 安全性增强：通过 Redis 支持多因素认证和操作限流，增强系统的安全性，防止暴力破解和恶意请求。
- 用户体验优化：缓存用户数据和会话管理，快速响应用户请求，优化整体用户体验