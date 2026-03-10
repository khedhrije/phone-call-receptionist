package helpers

import "golang.org/x/crypto/bcrypt"

// HashPassword generates a bcrypt hash of the given plain-text password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a plain-text password against a bcrypt hash.
// Returns true if they match.
func CheckPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
