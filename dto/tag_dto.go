package dto

import "my-api/models"

// CreateTagRequest represents request to create a new tag
type CreateTagRequest struct {
	Name  string `json:"name" binding:"required,max=50"`
	Color string `json:"color,omitempty"`
	Icon  string `json:"icon,omitempty"`
}

// UpdateTagRequest represents request to update a tag
type UpdateTagRequest struct {
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
	Icon  *string `json:"icon,omitempty"`
}

// AddTagsToTransactionRequest represents request to add tags to a transaction
type AddTagsToTransactionRequest struct {
	TagIDs []uint `json:"tag_ids" binding:"required,min=1"`
}

// TagSuggestion represents a suggested tag with confidence score
type TagSuggestion struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
}

// TagSpending represents spending data for a tag
type TagSpending struct {
	Tag              models.Tag `json:"tag"`
	TotalAmount      float64    `json:"total_amount"`
	TransactionCount int        `json:"transaction_count"`
	AvgAmount        float64    `json:"avg_amount"`
}

// TagSpendingResponse represents the response for spending by tag analytics
type TagSpendingResponse struct {
	Data   []TagSpending `json:"data"`
	Period PeriodInfo    `json:"period"`
}

// PeriodInfo represents the time period for analytics
type PeriodInfo struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}
