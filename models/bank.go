package models

type Bank struct {
    ID       uint   `json:"id" gorm:"primaryKey;type:int unsigned"`
    BankName string `json:"bank_name"`
    Color    string `json:"color"`
    Image    string `json:"image"`
}
