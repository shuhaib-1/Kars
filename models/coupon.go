package models

import (
	"time"
	"gorm.io/gorm"
)

type Coupon struct {
	gorm.Model
	CouponName      string    `json:"coupon_name" gorm:"not null"`
	CouponCode      string    `json:"coupon_code" gorm:"unique;not null"`
	DiscountType    string    `json:"discount_type" gorm:"not null"`
	DiscountValue   float64   `json:"discount_value" gorm:"type:decimal(10,2);not null"`
	MaximumDiscount float64   `json:"maximum_discount" gorm:"type:decimal(10,2)"`
	MinimumAmount   float64   `json:"minimum_amount" gorm:"type:decimal(10,2)"`
	UsageLimit      int       `json:"usage_limit" gorm:"default:0"`
	StartDate       time.Time `json:"start_date" gorm:"not null"`
	ExpiryTime      time.Time `json:"expiry_time" gorm:"not null"`
	IsActive        bool      `json:"is_active" gorm:"not null"`
}

type CouponUsage struct {
	gorm.Model
	CouponCode string `json:"coupon_code" gorm:"not null"`
	UserID   uint `json:"user_id" gorm:"not null"`
	Limit    int  `json:"usage" gorm:"default:0"`
}