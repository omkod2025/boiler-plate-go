package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"slices"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

// Config สำหรับกำหนดค่า JWT
type Config struct {
	PrivateKey       *rsa.PrivateKey
	PublicKey        *rsa.PublicKey
	TokenDuration    time.Duration
	Issuer           string
	Audience         string
	ValidateAudience bool // flag สำหรับควบคุมการตรวจสอบ audience
}

// Claims สำหรับ JWT payload
type Claims struct {
	UserID   string `json:"sub"`
	UserCode string `json:"subCode"`
	Role     string `json:"role"`
	gojwt.RegisteredClaims
}

// GenerateToken สร้าง JWT token แบบ RS256
func GenerateToken(config Config, userID, userCode, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		UserCode: userCode,
		Role:     role,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(now.Add(config.TokenDuration)),
			IssuedAt:  gojwt.NewNumericDate(now),
			NotBefore: gojwt.NewNumericDate(now),
			Issuer:    config.Issuer,
		},
	}

	// เพิ่ม Audience เฉพาะเมื่อต้องการ
	if config.Audience != "" {
		claims.Audience = []string{config.Audience}
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims)
	return token.SignedString(config.PrivateKey)
}

// ParseToken validate JWT token และ return claims
func ParseToken(tokenString string, publicKey *rsa.PublicKey, validateAudience bool, expectedAudience string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenString, &Claims{}, func(token *gojwt.Token) (interface{}, error) {
		// ตรวจสอบ signing method
		if _, ok := token.Method.(*gojwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	// ตรวจสอบ audience เฉพาะเมื่อต้องการ
	if validateAudience && expectedAudience != "" && !slices.Contains(claims.Audience, expectedAudience) {
		return nil, errors.New("invalid audience")
	}
	return claims, nil
}
