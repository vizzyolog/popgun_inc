package model

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

var KnownRoles = []string{"admin", "developer", "accounter", "repairman"}

type LoginForm struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
	Role     string `form:"role" json:"role"`
}

type User struct {
	gorm.Model   `json:"-"`
	PublicId     string    `gorm:"string" json:"uid"`
	UserName     string    `gorm:"unique" json:"username"`
	Password     string    `json:"-"`
	PasswordHash string    `json:"-"`
	PasswordSalt string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAT    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAT    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func CalculatePasswordHash(pass string) (passwordSalt string, passwordHash string, err error) {
	if pass == "" {
		err = errors.New("password must be set")
		return
	}
	passwordSalt = generateRandomString(10)
	passwordHash = hashSHA256([]byte(pass + passwordSalt))
	return
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func hashSHA256(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

func CheckPassword(password, passSalt, passHash string) bool {
	if passHash == "" {
		return false
	}
	hash := hashSHA256([]byte(password + passSalt))
	return hash == passHash
}
