package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("invalid password hash format")
	ErrIncompatibleVersion = errors.New("incompatible argon2 version")
)

// PasswordConfig holds password hashing configuration
type PasswordConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultPasswordConfig returns the default password configuration
func DefaultPasswordConfig() *PasswordConfig {
	return &PasswordConfig{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// PasswordService handles password hashing and verification
type PasswordService struct {
	config *PasswordConfig
}

// NewPasswordService creates a new password service
func NewPasswordService(config *PasswordConfig) *PasswordService {
	if config == nil {
		config = DefaultPasswordConfig()
	}
	return &PasswordService{config: config}
}

// HashPassword hashes a password using Argon2id
func (s *PasswordService) HashPassword(password string) (string, error) {
	salt := make([]byte, s.config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		s.config.Iterations,
		s.config.Memory,
		s.config.Parallelism,
		s.config.KeyLength,
	)

	// Encode as: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		s.config.Memory,
		s.config.Iterations,
		s.config.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// VerifyPassword verifies a password against a hash
func (s *PasswordService) VerifyPassword(password, encodedHash string) (bool, error) {
	config, salt, hash, err := s.decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Compute hash with same parameters
	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Iterations,
		config.Memory,
		config.Parallelism,
		config.KeyLength,
	)

	// Constant-time comparison
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

// decodeHash extracts the parameters, salt, and hash from an encoded hash
func (s *PasswordService) decodeHash(encodedHash string) (*PasswordConfig, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	config := &PasswordConfig{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &config.Memory, &config.Iterations, &config.Parallelism)
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	config.SaltLength = uint32(len(salt))

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	config.KeyLength = uint32(len(hash))

	return config, salt, hash, nil
}

// NeedsRehash checks if a hash needs to be rehashed with current config
func (s *PasswordService) NeedsRehash(encodedHash string) bool {
	config, _, _, err := s.decodeHash(encodedHash)
	if err != nil {
		return true
	}

	return config.Memory != s.config.Memory ||
		config.Iterations != s.config.Iterations ||
		config.Parallelism != s.config.Parallelism ||
		config.KeyLength != s.config.KeyLength
}
