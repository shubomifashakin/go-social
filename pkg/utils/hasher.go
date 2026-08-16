package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func HashPassword(password string) (string, error) {
    // generate random salt
    salt := make([]byte, 16)
    if _, err := rand.Read(salt); err != nil {
        return "", err
    }

    // hash the password
    hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)

    // format: $argon2id$salt$hash (both base64 encoded)
    encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
    encodedHash := base64.RawStdEncoding.EncodeToString(hash)

    return fmt.Sprintf("$argon2id$%s$%s", encodedSalt, encodedHash), nil
}

func VerifyPassword(password, storedHash string) bool {
    parts := strings.Split(storedHash, "$")
    // parts[0] = "", parts[1] = "argon2id", parts[2] = salt, parts[3] = hash
    
    salt, _ := base64.RawStdEncoding.DecodeString(parts[2])
    expectedHash, _ := base64.RawStdEncoding.DecodeString(parts[3])
    
    newHash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
    
    return subtle.ConstantTimeCompare(newHash, expectedHash) == 1
}