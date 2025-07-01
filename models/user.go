package models

import "my-api/config"

type User struct {
    ID    uint   `json:"id" gorm:"primaryKey"`
    Name  string `json:"name"`
    Email string `json:"email" gorm:"unique"`
	Address string `json:"address"`
    isVerified bool   `json:"is_verified" gorm:"default:false"`
    IsAdmin bool   `json:"is_admin" gorm:"default:false"`
    Password string `json:"password" gorm:"not null"`
}

func AutoMigrate() {
    config.DB.AutoMigrate(&User{})
}
