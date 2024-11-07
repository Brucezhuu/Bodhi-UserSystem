# Bodhi Usersystem Database Schema

## 1. `users` Table: Extending User Information
The `users` table stores basic user information with additional fields to support more comprehensive user management functions.

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    email TEXT UNIQUE,
    phone TEXT UNIQUE,
    account_status TEXT DEFAULT 'active',          -- User account status
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP                           -- For soft deletion
);
```

- `username` and `password`: Used for authentication.
- `email` and `phone`: Supports multiple contact methods.
- `account_status`: Indicates account status (e.g., active, suspended, disabled) for better account management.
- `deleted_at`: Used for soft deletion, making data recovery easier.

## 2. `roles` and `user_roles` Tables: Supporting User Roles and Permissions Management
The `roles` table defines roles in the system, while `user_roles` associates users with roles, supporting a many-to-many relationship to implement basic role-based access control (RBAC).
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
- The `roles` table defines roles within the system.
- The `user_roles` table establishes a many-to-many relationship between users and roles for assigning user permissions.

## 3. `user_activity_logs` Table: User Activity Logs
Logs user activities like login, logout, and password changes, making it easier to audit and track activity to enhance system security.
```sql
CREATE TABLE user_activity_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    activity_type TEXT NOT NULL,                   -- Activity type, e.g., login, logout
    activity_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,                               -- Records user IP address
    user_agent TEXT,                               -- User device information
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

```
- `activity_type`: Records activity type (e.g., login, logout).
- `ip_address` and `user_agent`: Provides information about user devices and access points for auditing purposes.

## 4. `user_mfa` Table: Multi-Factor Authentication (MFA) Support
Supports multi-factor authentication (MFA) by storing temporary OTP codes in Redis and configuring them in the database to enhance account security.
```sql
CREATE TABLE user_mfa (
    user_id INTEGER PRIMARY KEY,
    mfa_type TEXT NOT NULL,                        -- MFA type, e.g., SMS, email, TOTP
    mfa_secret TEXT NOT NULL,                      -- MFA secret or code
    is_enabled BOOLEAN DEFAULT FALSE,              -- Indicates if MFA is enabled
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

```

- `mfa_type`: Supports various MFA types.
- `mfa_secret`: Stores the secret or code for MFA.
- `is_enabled`: Indicates whether MFA is enabled.
## 5. `user_preferences` Table: User Preferences
Supports user customization for interface and notification settings to improve user experience.
```sql
CREATE TABLE user_preferences (
    user_id INTEGER PRIMARY KEY,
    theme TEXT DEFAULT 'light',                    -- Interface theme
    language TEXT DEFAULT 'en',                    -- Language preference
    receive_notifications BOOLEAN DEFAULT TRUE,    -- Notification preferences
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);


```
- `theme` and `language`: Supports customization of interface and language settings.
- `receive_notifications`:Allows users to manage notification preferences.

---

## Cache and Database Architecture with Redis
The introduction of Redis adds a caching layer, significantly improving system performance, response speed, and scalability, especially in areas like user management, access control, and rate limiting.

### Redis Use Cases

#### 1. User Session Management
Stores user session information (e.g., JWT) in Redis, enabling efficient user identity verification and enhancing both user experience and system performance.

#### 2. Caching User Information
Caches frequently accessed user data in Redis to reduce database queries and decrease latency. For example, in the `GetUser` operation, it first attempts to retrieve user information from Redis, and if the cache misses, it queries the database and stores the result in the cache.
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
    // Database query logic omitted
}

```
#### 3. Limiting Frequent Operations (Rate Limiting)
Uses Redis counters to limit the frequency of user operations. For instance, during login, it allows a maximum of 5 login attempts within 1 minute, blocking further login requests if exceeded. This method helps prevent brute-force attacks and abuse of interfaces.
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
    // Login logic omitted
}

```

#### 4. Storing Temporary Data for Multi-Factor Authentication (MFA)
Stores OTP and similar temporary codes in Redis to support MFA. A short expiration time (TTL) ensures the security and reliability of the codes.

```go
func GenerateMFA(userID int) string {
    otp := generateOTP() // Generate a one-time password
    otpKey := fmt.Sprintf("mfa:%d", userID)
    cache.RedisClient.Set(ctx, otpKey, otp, 5*time.Minute) // Valid for 5 minutes
    return otp
}

```
#### 5. Caching User Activity Logs
For frequently accessed activity records, Redis caches recent activity logs, allowing quick retrieval of recent user activities and reducing database load.

### Benefits of Redis
- **Performance Improvement**: Redis as a caching layer significantly reduces database access frequency, improving response speed.
- **Scalability**: Caching user information, MFA, and rate limiting mechanisms allow the system to support higher concurrency and more complex business requirements.
- **Enhanced Security**: Redis supports multi-factor authentication and rate limiting, enhancing system security by preventing brute-force attacks and malicious requests.
- **Improved User Experience**: By caching user data and managing sessions, Redis allows for faster user request responses, optimizing the overall user experience.