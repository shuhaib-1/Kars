package models

import (
	"gorm.io/gorm"
)

type Category struct {
	gorm.Model
	CategoryName string    `gorm:"type:varchar(255);not null" json:"category_name"`
	IsListed     string    `gorm:"type:varchar(10);default:'listed'" json:"is_listed"`
	OfferType    string    `gorm:"type:varchar(10);default:null" json:"offer_type"`
	OfferValue   float64   `gorm:"type:decimal;default:0" json:"offer_value"`
	Products     []Product `gorm:"foreignKey:CategoryID" json:"products"`
}
