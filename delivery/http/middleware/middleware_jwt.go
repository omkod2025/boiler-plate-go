package middleware

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	pkgjwt "github.com/omkod2025-boop/omgon-notification-service/pkg/jwt"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/logger"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/response"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTConfig สำหรับกำหนดค่า JWT (alias ของ pkg/jwt.Config)
type JWTConfig = pkgjwt.Config

// Claims สำหรับ JWT payload (alias ของ pkg/jwt.Claims)
type Claims = pkgjwt.Claims

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
	return pkgjwt.GenerateToken(config, userID, userCode, role)
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
	return pkgjwt.ParseToken(tokenString, publicKey, validateAudience, expectedAudience)
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
