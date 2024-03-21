package model

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	gorm.Model  `json:"-"`
	ID          string `json:"id"`
	Description string `json:"description"`
	Assignee    string `json:"assignee"`
	Status      bool   `json:"status"`
}

type User struct {
	gorm.Model `json:"-"`
	PublicId   string    `gorm:"string" json:"uid"`
	UserName   string    `gorm:"unique" json:"username"`
	Role       string    `json:"role"`
	CreatedAT  time.Time `json:"created_at"`
	UpdatedAT  time.Time `json:"updated_at"`
}
