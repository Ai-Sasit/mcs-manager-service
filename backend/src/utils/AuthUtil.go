package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

// HashPassword hashes a plaintext password using Argon2id.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Time, argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}

// CheckPasswordHash verifies a password against a stored hash.
// Returns an error if the password is wrong, the hash format is unrecognised,
// or the stored hash is the old bcrypt format (prompts reset).
func CheckPasswordHash(password, encoded string) error {
	// Support legacy bcrypt hashes from older versions of the application.
	if strings.HasPrefix(encoded, "$2a$") || strings.HasPrefix(encoded, "$2b$") {
		return bcrypt.CompareHashAndPassword([]byte(encoded), []byte(password))
	}

	parts := strings.Split(encoded, "$")
	// Expected: ["", "argon2id", "v=...", "m=...,t=...,p=...", "<salt>", "<hash>"]
	if len(parts) != 6 || parts[1] != "argon2id" {
		return fmt.Errorf("unsupported hash format")
	}

	var mem, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &time, &threads); err != nil {
		return fmt.Errorf("invalid hash parameters")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("invalid salt encoding")
	}

	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fmt.Errorf("invalid hash encoding")
	}

	computed := argon2.IDKey([]byte(password), salt, time, mem, threads, uint32(len(storedHash)))
	if subtle.ConstantTimeCompare(computed, storedHash) != 1 {
		return fmt.Errorf("invalid password")
	}
	return nil
}
