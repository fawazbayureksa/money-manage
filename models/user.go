package models

import "my-api/config"

type User struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	Name       string `json:"name"`
	Email      string `json:"email" gorm:"unique"`
	Address    string `json:"address"`
	IsVerified bool   `json:"is_verified" gorm:"default:false"`
	IsAdmin    bool   `json:"is_admin" gorm:"default:false"`
	Password   string `json:"-" gorm:"not null"` // Never expose password in JSON
}

func AutoMigrate() {
	config.DB.AutoMigrate(&User{}, &Bank{}, &Category{}, &Transaction{})
}
