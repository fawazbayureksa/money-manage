package dto

type BankResponse struct {
	ID       uint   `json:"id"`
	BankName string `json:"bank_name"`
	Color    string `json:"color"`
	Image    string `json:"image"`
}

type CreateBankRequest struct {
	BankName string `json:"bank_name" binding:"required"`
	Color    string `json:"color" binding:"required"`
	Image    string `json:"image"`
}

type BankFilterRequest struct {
	PaginationRequest
	BankName string `form:"bank_name"`
	Color    string `form:"color"`
}
