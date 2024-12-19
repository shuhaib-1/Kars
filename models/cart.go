package models

import (
	"gorm.io/gorm"
)

type Cart struct {
	gorm.Model            // Adds fields ID, CreatedAt, UpdatedAt, DeletedAt
	UserID     uint       `json:"user_id"` // Foreign key reference to User
	TotalItems int        `json:"total_items"`
	CartItems  []CartItem `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE;"` // One-to-many relationship with CartItem
}

type CartItem struct {
	gorm.Model           // Adds fields ID, CreatedAt, UpdatedAt, DeletedAt
	CartID       uint    `json:"cart_id"`    // Foreign key reference to Cart
	ProductID    uint    `json:"product_id"` // Reference to the product ID
	ProductName  string  `json:"product_name"`
	ProductPrice float64 `json:"product_price"`
	TotalPrice   float64 `json:"total_price"` // Using float64 for decimal values
	Quantity     int     `json:"quantity"`
}
