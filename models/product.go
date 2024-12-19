package models

import (
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	ProductName string   `gorm:"type:varchar(255);not null" json:"product_name"`
	Description string   `gorm:"type:text" json:"description"`
	Price       float64  `gorm:"type:decimal;not null" json:"price"`
	Quantity    int      `gorm:"not null;default:0" json:"quantity"`
	CategoryID  uint     `gorm:"not null" json:"category_id"`
	Category    Category `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"category_name"`
	Color       string   `gorm:"type:varchar(50)" json:"color"`
	ImgURLs     string   `gorm:"type:text" json:"img_urls"`
	Status      string   `gorm:"type:varchar(20)" json:"status"`
	OfferType   string   `gorm:"type:varchar(20);default:null" json:"offer_type"`
	OfferValue  float64  `gorm:"type:decimal;default:0" json:"offer_value"`
	IsListed    string   `gorm:"type:varchar(20);default:'listed'" json:"is_listed"`
}
