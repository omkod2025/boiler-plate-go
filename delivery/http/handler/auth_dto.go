package handler

import (
	"github.com/omkod2025/boiler-plate-go/domain/auth"
)

// Request DTO
type LoginRequest struct {
	Email    string `json:"email" th:"อีเมล" binding:"required,email,max=255"`
	Password string `json:"password" th:"รหัสผ่าน" binding:"required,max=72"`
}

func (r LoginRequest) toInput() auth.LoginInput {
	return auth.LoginInput{Email: r.Email, Password: r.Password}
}

// Response DTO
type TokenResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int    `json:"expiresIn"` // วินาที
}

func toTokenResponse(t auth.Token) TokenResponse {
	return TokenResponse{
		AccessToken: t.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(t.ExpiresIn.Seconds()),
	}
}
