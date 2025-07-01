package models

// import "my-api/config"

type Bank struct {
    ID    uint   `json:"id" gorm:"primaryKey"`
    BankName  string `json:"bank_name"`
    Color string `json:"color"`
	Image string `json:"image"`
}

// func AutoMigrate() {
//     config.DB.AutoMigrate(&Bank{})
// }
