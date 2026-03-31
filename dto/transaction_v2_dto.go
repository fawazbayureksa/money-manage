package dto

import (
	"my-api/models"
	"my-api/utils"
)

// SplitItem represents a single category split within a transaction request.
type SplitItem struct {
	CategoryID  *uint  `json:"category_id"`
	Amount      int    `json:"amount" binding:"required,min=1"`
	Description string `json:"description"`
}

// SplitItemResponse represents a single category split in a transaction response.
type SplitItemResponse struct {
	ID           uint   `json:"id"`
	CategoryID   *uint  `json:"category_id"`
	CategoryName string `json:"category_name"`
	Amount       int    `json:"amount"`
	Description  string `json:"description"`
}

// TransactionV2Response represents transaction response with asset information
type TransactionV2Response struct {
	ID              uint                `json:"id"`
	Description     string              `json:"description"`
	Amount          int                 `json:"amount"`
	TransactionType int                 `json:"transaction_type"`
	Date            utils.CustomTime    `json:"date"`
	CategoryName    string              `json:"category_name"`
	BankName        string              `json:"bank_name,omitempty"`
	AssetID         uint64              `json:"asset_id"`
	AssetName       string              `json:"asset_name,omitempty"`
	AssetType       string              `json:"asset_type,omitempty"`
	AssetBalance    float64             `json:"asset_balance,omitempty"`
	AssetCurrency   string              `json:"asset_currency,omitempty"`
	Tags            []models.Tag        `json:"tags,omitempty"`
	Splits          []SplitItemResponse `json:"splits,omitempty"`
}

// CreateTransactionV2Request represents request to create transaction with asset.
// When Splits is provided, CategoryID on the parent is optional and the split
// amounts must sum to Amount. When Splits is empty, CategoryID is required.
type CreateTransactionV2Request struct {
	Description     string      `json:"description" binding:"required"`
	CategoryID      *uint       `json:"category_id"`
	AssetID         uint64      `json:"asset_id" binding:"required"`
	Amount          int         `json:"amount" binding:"required,min=1"`
	TransactionType string      `json:"transaction_type" binding:"required,oneof=Income Expense income expense"`
	Date            string      `json:"date" binding:"required"`
	TagIDs          []uint      `json:"tag_ids,omitempty"`
	Splits          []SplitItem `json:"splits,omitempty"`
}

// UpdateTransactionV2Request represents request to update transaction.
// When Splits is non-nil the existing splits are fully replaced with the new list.
type UpdateTransactionV2Request struct {
	Description     *string      `json:"description"`
	CategoryID      *uint        `json:"category_id"`
	AssetID         *uint64      `json:"asset_id"`
	Amount          *int         `json:"amount"`
	TransactionType *string      `json:"transaction_type"`
	Date            *string      `json:"date"`
	TagIDs          *[]uint      `json:"tag_ids,omitempty"`
	Splits          *[]SplitItem `json:"splits,omitempty"`
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
