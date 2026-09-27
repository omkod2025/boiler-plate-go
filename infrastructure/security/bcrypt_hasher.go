package security

import "golang.org/x/crypto/bcrypt"

// BcryptHasher implements users.PasswordHasher และ auth.PasswordVerifier ด้วย bcrypt
type BcryptHasher struct {
	cost int
}

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{cost: bcrypt.DefaultCost}
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	return string(hash), err
}

func (h *BcryptHasher) Verify(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
