package models

import "time"

type Customer struct {
	CustomerID uint       `gorm:"primaryKey" json:"id"`
	Name       string     `gorm:"size:100;not null" json:"name" binding:"required"`
	Email      string     `gorm:"uniqueIndex;size:100;not null" json:"email" binding:"required,email,unique_customer_email"`
	Phone      string     `gorm:"size:20;not null" json:"phone" binding:"required,phone_id"`
	Address    string     `json:"address"`
	Orders     []Order    `gorm:"foreignKey:CustomerID" json:"orders,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	CreatedBy  string     `gorm:"default:'SYSTEM';not null" json:"-"`
	UpdatedAt  time.Time  `json:"-"`
	UpdatedBy  string     `gorm:"default:'SYSTEM';not null" json:"-"`
	DeletedAt  *time.Time `gorm:"column:deleted_at" json:"deleteAt,omitempty"`
	DeletedBy  *string    `gorm:"column:deleted_by;size:100" json:"deleteBy,omitempty"`
}

type CreateCustomerRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email,unique_customer_email"`
	Phone   string `json:"phone" binding:"required,phone_id"`
	Address string `json:"address"`
}

type UpdateCustomerRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Phone   string `json:"phone" binding:"required,phone_id"`
	Address string `json:"address"`
}

type CustomerOptionResponse struct {
	CustomerID uint   `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
}
