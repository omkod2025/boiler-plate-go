package middleware

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/logger"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/response"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig สำหรับกำหนดค่า JWT
type JWTConfig struct {
	PrivateKey       *rsa.PrivateKey
	PublicKey        *rsa.PublicKey
	TokenDuration    time.Duration
	Issuer           string
	Audience         string
	ValidateAudience bool // เพิ่ม flag สำหรับควบคุมการตรวจสอบ audience
}

// Claims สำหรับ JWT payload
type Claims struct {
	UserID   string `json:"sub"`
	UserCode string `json:"subCode"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWT middleware สำหรับ validate token
func JWT(config JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ดึง token จาก header
		tokenString, err := extractToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "unauthorized",
				"message": "Token not provided",
			})
			return
		}

		// Validate token
		claims, err := validateToken(tokenString, config.PublicKey, config.ValidateAudience, config.Audience)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "unauthorized",
				"message": "Invalid token",
			})
			return
		}

		// เก็บ claims ไว้ใน context เพื่อใช้ใน handler
		c.Set("userId", claims.UserID)
		c.Set("userCode", claims.UserCode)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// JWTOptional middleware สำหรับ validate token แบบ optional (ไม่บังคับ)
func JWTOptional(config JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractToken(c)
		if err != nil {
			response.Unauthorized(c, "token not found")
			c.Abort()
			return
		}

		claims, err := validateToken(tokenString, config.PublicKey, config.ValidateAudience, config.Audience)
		if err != nil {
			response.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}

		logger.Info("claims", claims)

		// เก็บ claims ไว้ใน context
		c.Set("userId", claims.UserID)
		c.Set("userCode", claims.UserCode)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// JWTRole middleware สำหรับตรวจสอบ role
func JWTRole(config JWTConfig, requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ใช้ JWT middleware ก่อน
		JWT(config)(c)

		// ถ้า JWT middleware abort แล้ว ให้ return ออก
		if c.IsAborted() {
			return
		}

		// ตรวจสอบ role
		userRole, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  "forbidden",
				"message": "Role not found in token",
			})
			return
		}

		role := userRole.(string)
		hasRole := false
		for _, requiredRole := range requiredRoles {
			if role == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  "forbidden",
				"message": fmt.Sprintf("Required roles: %v", requiredRoles),
			})
			return
		}

		c.Next()
	}
}

// GenerateToken สำหรับสร้าง JWT token
func GenerateToken(config JWTConfig, userID, userCode, role string) (string, error) {
	claims := Claims{
		UserID:   userID,
		UserCode: userCode,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.TokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    config.Issuer,
			// ไม่ใส่ Audience ถ้าไม่ต้องการตรวจสอบ
		},
	}

	// เพิ่ม Audience เฉพาะเมื่อต้องการ
	if config.Audience != "" {
		claims.Audience = []string{config.Audience}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(config.PrivateKey)
}

// ValidateToken สำหรับ validate token แยกออกมา
func ValidateToken(tokenString string, publicKey *rsa.PublicKey) (*Claims, error) {
	return validateToken(tokenString, publicKey, false, "")
}

// extractToken ดึง token จาก Authorization header
func extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header not found")
	}

	// ตรวจสอบ format "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// validateToken validate JWT token และ return claims
func validateToken(tokenString string, publicKey *rsa.PublicKey, validateAudience bool, expectedAudience string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// ตรวจสอบ signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// ตรวจสอบ audience เฉพาะเมื่อต้องการ
		if validateAudience && expectedAudience != "" {
			if !contains(claims.Audience, expectedAudience) {
				return nil, errors.New("invalid audience")
			}
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// contains ตรวจสอบว่า slice มีค่าที่ต้องการหรือไม่
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetUserFromContext helper function สำหรับดึง user info จาก context
func GetUserFromContext(c *gin.Context) (userID, userCode, role string, exists bool) {
	userIDValue, exists := c.Get("userId")
	if !exists {
		return
	}

	userID, ok := userIDValue.(string)
	if !ok {
		return
	}

	userCodeValue, _ := c.Get("userCode")
	userCode, _ = userCodeValue.(string)

	roleValue, _ := c.Get("role")
	role, _ = roleValue.(string)

	return userID, userCode, role, true
}

// GetClaimsFromContext helper function สำหรับดึง claims จาก context
func GetClaimsFromContext(c *gin.Context) (*Claims, bool) {
	claimsValue, exists := c.Get("claims")
	if !exists {
		return nil, false
	}

	claims, ok := claimsValue.(*Claims)
	return claims, ok
}

// LoadRSAPrivateKey โหลด RSA private key จาก PEM string
func LoadRSAPrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// ลอง parse PKCS8 format
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %v", err)
		}

		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("key is not RSA private key")
		}
		return rsaKey, nil
	}

	return privateKey, nil
}

// LoadRSAPublicKey โหลด RSA public key จาก PEM string
func LoadRSAPublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	rsaKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("key is not RSA public key")
	}

	return rsaKey, nil
}

// GenerateRSAKeyPair สร้าง RSA key pair สำหรับทดสอบ
func GenerateRSAKeyPair() (privateKeyPEM, publicKeyPEM string, err error) {
	privateKey, err := rsa.GenerateKey(nil, 2048)
	if err != nil {
		return "", "", err
	}

	// สร้าง private key PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}))

	// สร้าง public key PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	publicKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}))

	return privateKeyPEM, publicKeyPEM, nil
}

// IgnoreRule สำหรับกำหนด method และ path/regex ที่จะ ignore
// Method: "GET", "POST", "PUT", "DELETE", "ANY" (ไม่สนใจ method)
// Pattern: path หรือ regex
// Example: {Method: "POST", Pattern: `^/[0-9]+/customer$`}
type IgnoreRule struct {
	Method  string
	Pattern string
}

// JWTWithIgnoreRules middleware สำหรับ validate token และ ignore ตาม method+path/regex
func JWTWithIgnoreRules(config JWTConfig, ignoreRules []IgnoreRule) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldIgnoreRequest(c.Request.Method, c.Request.URL.Path, ignoreRules) {
			c.Next()
			return
		}
		JWT(config)(c)
	}
}

// JWTOptionalWithIgnoreRules middleware สำหรับ validate token แบบ optional และ ignore ตาม method+path/regex
func JWTOptionalWithIgnoreRules(config JWTConfig, ignoreRules []IgnoreRule) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldIgnoreRequest(c.Request.Method, c.Request.URL.Path, ignoreRules) {
			c.Next()
			return
		}
		JWTOptional(config)(c)
	}
}

// shouldIgnoreRequest ตรวจสอบว่า method+path นี้ควร ignore หรือไม่
func shouldIgnoreRequest(method, path string, ignoreRules []IgnoreRule) bool {
	for _, rule := range ignoreRules {
		if rule.Method != "ANY" && !strings.EqualFold(rule.Method, method) {
			continue
		}
		// รองรับ regex pattern
		if len(rule.Pattern) > 0 && (rule.Pattern[0] == '^' || strings.ContainsAny(rule.Pattern, ".*+?[]()$")) {
			matched, _ := regexp.MatchString(rule.Pattern, path)
			if matched {
				return true
			}
		} else if path == rule.Pattern || (strings.HasSuffix(rule.Pattern, "/") && strings.HasPrefix(path, rule.Pattern)) {
			return true
		}
	}
	return false
}
