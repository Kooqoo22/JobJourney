package security

import (
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func PasswordStrengthErrors(field, pw string) []utils.FieldError {
	var hasLetter, hasDigit bool
	for _, r := range pw {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	var out []utils.FieldError
	if len([]rune(pw)) < 8 {
		out = append(out, utils.FieldError{Field: field, Message: "must be at least 8 characters"})
	}
	if !hasLetter || !hasDigit {
		out = append(out, utils.FieldError{Field: field, Message: "must contain both letters and numbers"})
	}
	return out
}
