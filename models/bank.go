package models

type Bank struct {
    ID    uint   `json:"id" gorm:"primaryKey"`
    BankName  string `json:"bank_name"`
    Color string `json:"color"`
	Image string `json:"image"`
}
