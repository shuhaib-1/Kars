package models

import "gorm.io/gorm"

type Wishlist struct {
    gorm.Model
    UserID            uint    `json:"user_id" gorm:"not null"`                      
    ProductID         uint    `json:"product_id" gorm:"not null"`                   
    ProductDescription string  `json:"product_description" gorm:"type:text"`         
    ProductName       string  `json:"product_name" gorm:"not null"`                
    ProductPrice      float64 `json:"product_price" gorm:"not null"`                
}
