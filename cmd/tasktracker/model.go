package main

import "time"

type Task struct {
	createdBy  User
	owner      User
	createdAt  time.Time
	finishedAt time.Time
	done       bool
}

type User struct {
	PublicID string
	Email    string
	Fullname string
	Position string
	Active   bool
	UserRole Role
}

type Role int

const (
	admin Role = iota + 1
	manager
	accountant
	developer
)
