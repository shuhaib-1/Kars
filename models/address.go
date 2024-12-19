package models

import (
	"gorm.io/gorm"
)

type Address struct {
	gorm.Model          // Adds ID, CreatedAt, UpdatedAt, and DeletedAt fields
	UserID       uint   `gorm:"not null"` // Foreign key for User
	Name         string `gorm:"not null"` // Name associated with the address
	PhoneNo      string `gorm:"not null"` // Contact phone number
	AddressLine1 string `gorm:"not null"`
	AddressLine2 string
	City         string `gorm:"not null"`
	State        string `gorm:"not null"`
	PostalCode   string `gorm:"not null"`
	Country      string `gorm:"not null"`
	LandMark     string
	AddressType  string `gorm:"not null"` // Could be 'shipping', 'billing', etc.
}
