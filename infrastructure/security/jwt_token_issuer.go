package security

import (
	"strconv"

	"github.com/omkod2025/boiler-plate-go/domain/auth"
	"github.com/omkod2025/boiler-plate-go/pkg/jwt"
)

// JWTTokenIssuer implements auth.TokenIssuer ด้วย JWT แบบ RS256
type JWTTokenIssuer struct {
	config jwt.Config
}

var _ auth.TokenIssuer = (*JWTTokenIssuer)(nil)

func NewJWTTokenIssuer(config jwt.Config) *JWTTokenIssuer {
	return &JWTTokenIssuer{config: config}
}

func (i *JWTTokenIssuer) Issue(userID int, role string) (auth.Token, error) {
	token, err := jwt.GenerateToken(i.config, strconv.Itoa(userID), "", role)
	if err != nil {
		return auth.Token{}, err
	}
	return auth.Token{AccessToken: token, ExpiresIn: i.config.TokenDuration}, nil
}
