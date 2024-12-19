package models

import "gorm.io/gorm"

type OrderAddress struct {
	Name         string
	PhoneNo      string
	AddressLine1 string
	AddressLine2 string
	City         string
	State        string
	PostalCode   string
	Country      string
	LandMark     string
}

type Order struct {
	gorm.Model
	UserID         uint         `json:"user_id"`
	TotalPrice     float64      `json:"total_price"`
	CouponID       *uint        `json:"coupon_id" gorm:"default:null"`
	DiscountAmount float64      `json:"discount_amount"`
	ShippingAmount float64      `json:"shipping_amount"`
	FinalPrice     float64      `json:"final_price"`
	OrderAddress   OrderAddress `json:"order_address" gorm:"embedded;embeddedPrefix:address_"`
	PaymentMethod  string       `json:"payment_method"`
	PaymentStatus  string       `json:"payment_status"`
	OrderStatus    string       `json:"order_status"`
	CouponCode     string       `json:"coupon_code"`
	OrderItems     []OrderItem  `json:"order_items" gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE;"`
}

type OrderItem struct {
	gorm.Model
	OrderID      uint    `json:"order_id" gorm:"not null;index;constraint:OnDelete:CASCADE;"`
	ProductID    uint    `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ProductPrice float64 `json:"product_price"`
	Quantity     int     `json:"quantity"`
	IsCancelled  string  `gorm:"type:varchar(10);default:ordered" json:"is_cancelled"`
	TotalPrice   float64 `json:"total_price"`
}
