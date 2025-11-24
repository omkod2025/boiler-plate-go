# JWT Middleware Usage Guide (RSA Keypair)

## Overview
JWT middleware สำหรับ validate JWT tokens ใน Gin framework โดยใช้ RSA keypair สำหรับ signing และ validation แทนการใช้ secret key

## RSA Keypair Configuration

### 1. สร้าง RSA Keypair
```bash
# รัน script เพื่อสร้าง RSA keypair
go run scripts/generate_rsa_keys.go
```

หรือใช้ OpenSSL:
```bash
# สร้าง private key
openssl genrsa -out private_key.pem 2048

# สร้าง public key จาก private key
openssl rsa -in private_key.pem -pubout -out public_key.pem
```

### 2. Environment Variables
```env
JWT_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
-----END RSA PRIVATE KEY-----

JWT_PUBLIC_KEY=-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A...
-----END PUBLIC KEY-----

JWT_DURATION=24h                    # 24 hours
JWT_ISSUER=fmg-auth-api
JWT_AUDIENCE=                       # ไม่บังคับ (optional)
JWT_VALIDATE_AUDIENCE=false         # ปิดการตรวจสอบ audience เป็นค่าเริ่มต้น
```

### 3. JWT Duration Formats
```env
# ตัวอย่างการกำหนด JWT_DURATION ในรูปแบบต่างๆ

# Development (ยาว)
JWT_DURATION=24h                    # 24 hours
JWT_DURATION=7d                     # 7 days
JWT_DURATION=30d                    # 30 days

# Production (สั้น)
JWT_DURATION=1h                     # 1 hour
JWT_DURATION=30m                    # 30 minutes
JWT_DURATION=15m                    # 15 minutes

# Custom durations
JWT_DURATION=2h30m                  # 2 hours 30 minutes
JWT_DURATION=1d12h                  # 1 day 12 hours
JWT_DURATION=1h30m45s               # 1 hour 30 minutes 45 seconds
```

### 4. JWT Claims Structure
```go
type Claims struct {
    UserID   string `json:"sub"`        // Subject - User ID (ตาม JWT standard)
    UserCode string `json:"subCode"`     // User Code สำหรับ business logic
    Role     string `json:"role"`        // User Role
    jwt.RegisteredClaims
}
```

## Audience Validation Options

### 1. ไม่ตรวจสอบ Audience (Default)
```env
JWT_VALIDATE_AUDIENCE=false
JWT_AUDIENCE=
```

### 2. ตรวจสอบ Audience
```env
JWT_VALIDATE_AUDIENCE=true
JWT_AUDIENCE=fmg-auth-users
```

### 3. ตรวจสอบ Audience แบบ Multiple
```env
JWT_VALIDATE_AUDIENCE=true
JWT_AUDIENCE=fmg-auth-users,fmg-payment-api
```

## Middleware Types

### 1. JWT Middleware (Required)
```go
// ต้องมี JWT token ถึงจะเข้าถึงได้
protected := r.Group("/protected")
protected.Use(middleware.JWT(jwtConfig))
{
    protected.GET("/profile", handler.GetProfile)
}
```

### 2. JWTRole Middleware (Role-based)
```go
// ต้องมี JWT token และ role ที่กำหนด
admin := r.Group("/admin")
admin.Use(middleware.JWTRole(jwtConfig, "admin"))
{
    admin.GET("/users", handler.GetAllUsers)
}
```

### 3. JWTOptional Middleware (Optional)
```go
// มีหรือไม่มี JWT token ก็ได้
optional := r.Group("/optional")
optional.Use(middleware.JWTOptional(jwtConfig))
{
    optional.GET("/public", handler.GetPublicData)
}
```

## Helper Functions

### GetUserFromContext
```go
func (h *Handler) GetProfile(c *gin.Context) {
    userID, userCode, role, exists := middleware.GetUserFromContext(c)
    if !exists {
        response.Unauthorized(c, "User not found in context")
        return
    }
    
    // ใช้ userID, userCode, role
}
```

### GetClaimsFromContext
```go
func (h *Handler) GetFullClaims(c *gin.Context) {
    claims, exists := middleware.GetClaimsFromContext(c)
    if !exists {
        response.Unauthorized(c, "Claims not found")
        return
    }
    
    // ใช้ claims.UserID, claims.UserCode, claims.Role
}
```

## Token Generation

### GenerateToken
```go
token, err := middleware.GenerateToken(jwtConfig, "userId", "user_code", "role")
if err != nil {
    // handle error
}
```

## RSA Key Management

### LoadRSAPrivateKey
```go
privateKey, err := middleware.LoadRSAPrivateKey(privateKeyPEM)
if err != nil {
    // handle error
}
```

### LoadRSAPublicKey
```go
publicKey, err := middleware.LoadRSAPublicKey(publicKeyPEM)
if err != nil {
    // handle error
}
```

### GenerateRSAKeyPair (สำหรับทดสอบ)
```go
privateKeyPEM, publicKeyPEM, err := middleware.GenerateRSAKeyPair()
if err != nil {
    // handle error
}
```

## Example Usage

### Login Route
```go
func (h *Handler) Login(c *gin.Context) {
    // ตรวจสอบ credentials
    if username == "admin" && password == "admin123" {
        token, err := middleware.GenerateToken(h.JWTConfig, "1", "ADM001", "admin")
        if err != nil {
            response.Error(c, "Failed to generate token")
            return
        }
        
        response.Success(c, "Login successful", map[string]interface{}{
            "token": token,
            "user": map[string]interface{}{
                "user_id":   "1",
                "user_code": "ADM001",
                "role":      "admin",
            },
        })
    }
}
```

### Protected Route
```go
func (h *Handler) GetProfile(c *gin.Context) {
    userID, userCode, role, exists := middleware.GetUserFromContext(c)
    if !exists {
        response.Unauthorized(c, "User not found in context")
        return
    }
    
    profile := map[string]interface{}{
        "userId":   userID,
        "userCode": userCode,
        "role":      role,
    }
    
    response.Success(c, "Profile retrieved successfully", profile)
}
```

## API Testing

### 1. Login to get token
```bash
curl -X POST http://localhost:21400/api/example/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

### 2. Use token in protected routes
```bash
curl -X GET http://localhost:21400/api/example/protected/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 3. Admin routes (require admin role)
```bash
curl -X GET http://localhost:21400/api/example/admin/users \
  -H "Authorization: Bearer ADMIN_TOKEN_HERE"
```

### 4. Optional routes (with or without token)
```bash
# With token
curl -X GET http://localhost:21400/api/example/optional/public \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Without token
curl -X GET http://localhost:21400/api/example/optional/public
```

## Security Best Practices

### 1. Key Management
- **Private Key**: เก็บไว้ใน secure storage และไม่เปิดเผย
- **Public Key**: สามารถแชร์ได้สำหรับ validation
- **Key Rotation**: เปลี่ยน keys เป็นระยะ
- **Key Size**: ใช้ RSA 2048 bits ขึ้นไป

### 2. Environment Configuration
```bash
# ใช้ base64 encoding สำหรับ environment variables
JWT_PRIVATE_KEY=$(cat private_key.pem | base64 -w 0)
JWT_PUBLIC_KEY=$(cat public_key.pem | base64 -w 0)
```

### 3. Production Considerations
- ใช้ Hardware Security Module (HSM) สำหรับ private key
- ใช้ Key Management Service (KMS) เช่น AWS KMS, Azure Key Vault
- ตั้งเวลาหมดอายุที่เหมาะสม (1h สำหรับ production)
- ใช้ HTTPS เสมอ

### 4. Token Security
- เก็บ token ไว้ใน secure storage (ไม่ใช่ localStorage)
- ใช้ short-lived tokens
- Implement token refresh mechanism
- ใช้ secure cookie สำหรับ web applications

### 5. Duration Best Practices
- **Development**: `JWT_DURATION=24h` หรือ `JWT_DURATION=7d`
- **Production**: `JWT_DURATION=1h` หรือ `JWT_DURATION=30m`
- **High Security**: `JWT_DURATION=15m` หรือ `JWT_DURATION=5m`
- **Refresh Tokens**: `JWT_DURATION=7d` หรือ `JWT_DURATION=30d`

### 6. Audience Validation
- **Development**: ปิดการตรวจสอบ audience (`JWT_VALIDATE_AUDIENCE=false`)
- **Production**: เปิดการตรวจสอบ audience (`JWT_VALIDATE_AUDIENCE=true`)
- **Microservices**: ใช้ audience เพื่อจำกัด scope ของ token

## Duration Format Reference

| Format | Example | Description |
|--------|---------|-------------|
| `ns` | `1000000000ns` | nanoseconds |
| `us` | `1000000us` | microseconds |
| `ms` | `1000ms` | milliseconds |
| `s` | `60s` | seconds |
| `m` | `30m` | minutes |
| `h` | `2h` | hours |
| `d` | `7d` | days |

### Common Duration Examples
```env
# Short durations (Production)
JWT_DURATION=5m                    # 5 minutes
JWT_DURATION=15m                   # 15 minutes
JWT_DURATION=30m                   # 30 minutes
JWT_DURATION=1h                    # 1 hour

# Medium durations (Development)
JWT_DURATION=2h                    # 2 hours
JWT_DURATION=6h                    # 6 hours
JWT_DURATION=12h                   # 12 hours
JWT_DURATION=24h                   # 24 hours

# Long durations (Refresh tokens)
JWT_DURATION=7d                    # 7 days
JWT_DURATION=30d                   # 30 days
JWT_DURATION=90d                   # 90 days

# Complex durations
JWT_DURATION=1h30m                 # 1 hour 30 minutes
JWT_DURATION=2h15m30s              # 2 hours 15 minutes 30 seconds
JWT_DURATION=1d12h                 # 1 day 12 hours
```

## Advantages of RSA over HMAC

1. **Asymmetric**: Private key สำหรับ signing, public key สำหรับ validation
2. **Key Distribution**: Public key สามารถแชร์ได้อย่างปลอดภัย
3. **Microservices**: หลาย services สามารถ validate tokens ได้โดยไม่ต้องแชร์ secret
4. **Security**: ใช้ cryptographic algorithms ที่แข็งแกร่งกว่า
5. **Compliance**: ตรงตามมาตรฐานความปลอดภัยหลายมาตรฐาน
6. **Flexible Audience**: สามารถเปิด/ปิดการตรวจสอบ audience ได้ตามต้องการ
7. **Flexible Duration**: รองรับการกำหนด duration ในรูปแบบที่หลากหลาย 