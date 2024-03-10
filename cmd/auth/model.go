package main

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	PublicID     string
	Login        string
	Password     string
	PasswordHash string
	PasswordSalt string
	RoleID       UserRole
}

type UserRole int

const (
	Developer = iota
	Manager
	Accountant
	Admin
)

var AllRoles = []string{"Developer", "Manager", "Accountant", "Admin"}

func GetAllRoles() []string {

	return AllRoles
}

func generateNewID() string {
	return uuid.New().String()
}
