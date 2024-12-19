package models

import "gorm.io/gorm"

type Admin struct {
    gorm.Model
    AdminName string `gorm:"size:255;not null"`
    Email     string `gorm:"size:255;unique;not null"`
    Password  string `gorm:"size:255;not null"`
    Status   string `gorm:"type:varchar(20);default:'Inactive'"`
}