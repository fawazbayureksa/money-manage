package models

import "my-api/config"

type User struct {
    ID    uint   `json:"id" gorm:"primaryKey"`
    Name  string `json:"name"`
    Email string `json:"email" gorm:"unique"`
	Address string `json:"address"`
}

func AutoMigrate() {
    config.DB.AutoMigrate(&User{})
}
