package models

import "gorm.io/gorm"

type User struct {
    gorm.Model
    UserName string `gorm:"size:255"`
    Email    string `gorm:"size:255;not null;unique"`
    Password string `gorm:"size:255;not null"`
    PhoneNo  string `gorm:"size:10"` // Assuming phone numbers are stored as strings to avoid issues with leading zeros
    IsBlocked bool `gorm:"default:false"`
    Status   string `gorm:"type:varchar(20);default:'Inactive'"`
}
