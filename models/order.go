package models

import "time"

type Order struct {
	OrderID    uint        `gorm:"primaryKey;column:order_id" json:"orderId"`
	CustomerID uint        `gorm:"column:customer_id;not null" json:"-"`
	Customer   *Customer   `gorm:"foreignKey:CustomerID;references:CustomerID" json:"customer,omitempty"`
	OrderNo    string      `gorm:"column:order_no;uniqueIndex;size:50;not null" json:"orderNo"`
	Total      float64     `gorm:"column:total;type:numeric(15,2);not null;default:0.00" json:"total"`
	Status     string      `gorm:"column:status;size:20;not null;default:'PENDING'" json:"status"`
	Items      []OrderItem `gorm:"foreignKey:OrderID;references:OrderID" json:"items,omitempty"`
	CreatedAt  time.Time   `gorm:"column:created_at" json:"createdAt"`
	CreatedBy  string      `gorm:"column:created_by;not null;default:'SYSTEM'" json:"createdBy"`
	UpdatedAt  time.Time   `gorm:"column:updated_at" json:"updatedAt"`
	UpdatedBy  string      `gorm:"column:updated_by;not null;default:'SYSTEM'" json:"updatedBy"`
	DeletedAt  *time.Time  `gorm:"column:deleted_at" json:"deletedAt,omitempty"`
	DeletedBy  *string     `gorm:"column:deleted_by" json:"deletedBy,omitempty"`
}

type OrderItem struct {
	OrderItemID uint      `gorm:"primaryKey;column:order_item_id" json:"orderItemId"`
	OrderID     uint      `gorm:"column:order_id;not null" json:"-"`
	Name        string    `gorm:"column:name;size:50;not null" json:"name"`
	Qty         int       `gorm:"column:qty;not null" json:"qty"`
	Price       float64   `gorm:"column:price;type:numeric(15,2);not null;default:0.00" json:"price"`
	Total       float64   `gorm:"column:total;type:numeric(15,2);not null;default:0.00" json:"total"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
	CreatedBy   string    `gorm:"column:created_by;not null;default:'SYSTEM'" json:"createdBy"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt"`
	UpdatedBy   string    `gorm:"column:updated_by;not null;default:'SYSTEM'" json:"updatedBy"`
}

func (OrderItem) TableName() string {
	return "orders_item"
}

// --- Request DTOs ---

type OrderItemRequest struct {
	Name  string  `json:"name" binding:"required,max=50"`
	Qty   int     `json:"qty" binding:"required,gt=0"`
	Price float64 `json:"price" binding:"required,gt=0"`
}

type CreateOrderRequest struct {
	CustomerID uint               `json:"customerId" binding:"required"`
	Status     string             `json:"status" binding:"omitempty,oneof=PENDING PAID SHIPPED CANCELLED"`
	Items      []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type UpdateOrderRequest struct {
	CustomerID uint               `json:"customerId" binding:"required"`
	Status     string             `json:"status" binding:"required,oneof=PENDING PAID SHIPPED CANCELLED"`
	Items      []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
}
