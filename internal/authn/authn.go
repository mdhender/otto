// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package authn

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword uses bcrypt to hash the secret (from gregorygaines.com)
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(bytes), err
}
