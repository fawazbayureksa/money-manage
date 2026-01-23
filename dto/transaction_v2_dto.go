package dto

import (
	"my-api/utils"
)

// TransactionV2Response represents transaction response with asset information
type TransactionV2Response struct {
	ID              uint             `json:"id"`
	Description     string           `json:"description"`
	Amount          int              `json:"amount"`
	TransactionType int              `json:"transaction_type"`
	Date            utils.CustomTime `json:"date"`
	CategoryName    string           `json:"category_name"`
	BankName        string           `json:"bank_name,omitempty"`
	AssetID         uint64           `json:"asset_id"`
	AssetName       string           `json:"asset_name,omitempty"`
	AssetType       string           `json:"asset_type,omitempty"`
	AssetBalance    float64          `json:"asset_balance,omitempty"`
	AssetCurrency   string           `json:"asset_currency,omitempty"`
}

// CreateTransactionV2Request represents request to create transaction with asset
type CreateTransactionV2Request struct {
	Description     string `json:"description" binding:"required"`
	CategoryID      uint   `json:"category_id" binding:"required"`
	AssetID         uint64 `json:"asset_id" binding:"required"`
	Amount          int    `json:"amount" binding:"required,min=1"`
	TransactionType string `json:"transaction_type" binding:"required,oneof=Income Expense income expense"`
	Date            string `json:"date" binding:"required"`
}

// UpdateTransactionV2Request represents request to update transaction
type UpdateTransactionV2Request struct {
	Description     *string `json:"description"`
	CategoryID      *uint   `json:"category_id"`
	AssetID         *uint64 `json:"asset_id"`
	Amount          *int    `json:"amount"`
	TransactionType *string `json:"transaction_type"`
	Date            *string `json:"date"`
}

// AssetTransactionsResponse represents transactions for a specific asset
type AssetTransactionsResponse struct {
	AssetID        uint64                  `json:"asset_id"`
	AssetName      string                  `json:"asset_name"`
	AssetType      string                  `json:"asset_type"`
	CurrentBalance float64                 `json:"current_balance"`
	Currency       string                  `json:"currency"`
	Transactions   []TransactionV2Response `json:"transactions"`
	TotalIncome    float64                 `json:"total_income"`
	TotalExpense   float64                 `json:"total_expense"`
}
