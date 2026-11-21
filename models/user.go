package models

import "time"

type User struct {
	UserID       uint   `gorm:"primaryKey" json:"id"`
	Name         string `gorm:"uniqueIndex;size:50;not null" json:"name" binding:"required"`
	Email        string `gorm:"uniqueIndex;size:100;not null" json:"email" binding:"required,email"`
	PasswordHash string `gorm:"column:password;not null" json:"-"`
	Password     string `gorm:"-" json:"password,omitempty" binding:"required,min=6"`
	CreatedAt    time.Time
	CreatedBy    string `gorm:"default:'SYSTEM';not null" json:"-"`
	UpdatedAt    time.Time
	UpdatedBy    string `gorm:"default:'SYSTEM';not null" json:"-"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
