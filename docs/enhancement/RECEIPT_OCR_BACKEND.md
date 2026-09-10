# Receipt OCR Backend Implementation Plan

## Overview

Enable users to upload receipt/bill images and automatically extract transaction data using OCR (Optical Character Recognition) technology. This feature reduces manual data entry and improves accuracy.

---

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│  API Server │────▶│ OCR Service │────▶│   Parser    │
│  (Upload)   │     │  (Go/Gin)   │     │ (External)  │     │  (Extract)  │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                           │                                       │
                           ▼                                       ▼
                    ┌─────────────┐                         ┌─────────────┐
                    │   Storage   │                         │ Transaction │
                    │  (S3/Local) │                         │   (Create)  │
                    └─────────────┘                         └─────────────┘
```

---

## OCR Provider Options

### Option 1: Google Cloud Vision API (Recommended)
- **Pros**: High accuracy, multi-language support, structured data extraction
- **Cons**: Cost per request ($1.50 per 1000 images)
- **Best for**: Production environment

### Option 2: Tesseract OCR (Self-hosted)
- **Pros**: Free, no API limits, privacy-friendly
- **Cons**: Lower accuracy, requires server resources
- **Best for**: Development/testing or budget-conscious deployment

### Option 3: AWS Textract
- **Pros**: Good table/form extraction, AWS ecosystem integration
- **Cons**: Higher cost
- **Best for**: If already using AWS infrastructure

### Option 4: Azure Computer Vision
- **Pros**: Good accuracy, receipt-specific models
- **Cons**: Azure dependency
- **Best for**: If already using Azure

---

## Database Schema

### Table: `receipt_scans`

```sql
CREATE TABLE receipt_scans (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    
    -- Image storage
    original_image_url VARCHAR(500) NOT NULL,
    processed_image_url VARCHAR(500) NULL,
    
    -- OCR Results
    raw_ocr_text TEXT NULL,
    extracted_data JSON NULL COMMENT 'Parsed receipt data',
    confidence_score DECIMAL(5,2) NULL COMMENT '0-100%',
    
    -- Processing status
    status ENUM('uploaded', 'processing', 'completed', 'failed', 'reviewed') DEFAULT 'uploaded',
    error_message TEXT NULL,
    
    -- Linked transaction (after user confirms)
    transaction_id BIGINT UNSIGNED NULL,
    
    -- Metadata
    file_name VARCHAR(255) NULL,
    file_size INT NULL,
    mime_type VARCHAR(50) NULL,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_user_status (user_id, status),
    INDEX idx_transaction (transaction_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (transaction_id) REFERENCES transaction_v2(id) ON DELETE SET NULL
);
```

### Table: `receipt_templates`

Store merchant-specific parsing rules for improved accuracy.

```sql
CREATE TABLE receipt_templates (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    merchant_name VARCHAR(255) NOT NULL,
    merchant_patterns JSON NOT NULL COMMENT 'Regex patterns to identify merchant',
    
    -- Parsing rules
    date_patterns JSON NULL,
    total_patterns JSON NULL,
    item_patterns JSON NULL,
    
    -- Default category mapping
    default_category_id BIGINT UNSIGNED NULL,
    
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_merchant (merchant_name),
    FOREIGN KEY (default_category_id) REFERENCES categories(id) ON DELETE SET NULL
);
```

---

## API Endpoints

### 1. Upload Receipt Image

```http
POST /api/v2/receipts/scan
Authorization: Bearer {token}
Content-Type: multipart/form-data

Form Data:
- image: [binary file]
- asset_id: 1 (optional, pre-select wallet)
```

**Response:**
```json
{
    "success": true,
    "message": "Receipt uploaded and processing started",
    "data": {
        "scan_id": 123,
        "status": "processing",
        "image_url": "https://storage.example.com/receipts/abc123.jpg"
    }
}
```

### 2. Check Scan Status / Get Results

```http
GET /api/v2/receipts/scan/{scan_id}
Authorization: Bearer {token}
```

**Response (Processing):**
```json
{
    "success": true,
    "data": {
        "scan_id": 123,
        "status": "processing",
        "estimated_time": 5
    }
}
```

**Response (Completed):**
```json
{
    "success": true,
    "data": {
        "scan_id": 123,
        "status": "completed",
        "image_url": "https://storage.example.com/receipts/abc123.jpg",
        "extracted_data": {
            "merchant_name": "Indomaret",
            "date": "2026-02-11",
            "time": "14:35",
            "total_amount": 125000,
            "subtotal": 120000,
            "tax": 5000,
            "items": [
                {
                    "name": "Indomie Goreng",
                    "quantity": 5,
                    "unit_price": 3500,
                    "total": 17500
                },
                {
                    "name": "Aqua 600ml",
                    "quantity": 2,
                    "unit_price": 4000,
                    "total": 8000
                }
            ],
            "payment_method": "Cash",
            "suggested_category": {
                "id": 5,
                "name": "Groceries"
            }
        },
        "confidence_score": 87.5,
        "raw_text": "INDOMARET\nJl. Sudirman No. 123\n..."
    }
}
```

### 3. Confirm and Create Transaction

```http
POST /api/v2/receipts/scan/{scan_id}/confirm
Authorization: Bearer {token}

{
    "amount": 125000,
    "date": "2026-02-11",
    "description": "Indomaret - Groceries",
    "category_id": 5,
    "asset_id": 1,
    "transaction_type": 2,
    "notes": "Weekly groceries"
}
```

**Response:**
```json
{
    "success": true,
    "message": "Transaction created from receipt",
    "data": {
        "transaction_id": 456,
        "scan_id": 123
    }
}
```

### 4. List Receipt Scans

```http
GET /api/v2/receipts?status=completed&page=1&limit=10
Authorization: Bearer {token}
```

### 5. Delete Receipt Scan

```http
DELETE /api/v2/receipts/scan/{scan_id}
Authorization: Bearer {token}
```

---

## Go Implementation

### Models

```go
// models/receipt_scan.go

package models

import (
    "encoding/json"
    "time"
)

type ReceiptScan struct {
    ID                uint            `gorm:"primaryKey" json:"id"`
    UserID            uint            `gorm:"not null" json:"user_id"`
    OriginalImageURL  string          `gorm:"size:500;not null" json:"original_image_url"`
    ProcessedImageURL *string         `gorm:"size:500" json:"processed_image_url,omitempty"`
    RawOCRText        *string         `gorm:"type:text" json:"raw_ocr_text,omitempty"`
    ExtractedData     json.RawMessage `gorm:"type:json" json:"extracted_data,omitempty"`
    ConfidenceScore   *float64        `gorm:"type:decimal(5,2)" json:"confidence_score,omitempty"`
    Status            string          `gorm:"size:20;default:uploaded" json:"status"`
    ErrorMessage      *string         `gorm:"type:text" json:"error_message,omitempty"`
    TransactionID     *uint           `json:"transaction_id,omitempty"`
    FileName          *string         `gorm:"size:255" json:"file_name,omitempty"`
    FileSize          *int            `json:"file_size,omitempty"`
    MimeType          *string         `gorm:"size:50" json:"mime_type,omitempty"`
    CreatedAt         time.Time       `json:"created_at"`
    UpdatedAt         time.Time       `json:"updated_at"`
    
    // Relations
    User        User           `gorm:"foreignKey:UserID" json:"-"`
    Transaction *TransactionV2 `gorm:"foreignKey:TransactionID" json:"transaction,omitempty"`
}

type ExtractedReceiptData struct {
    MerchantName      string              `json:"merchant_name"`
    Date              string              `json:"date"`
    Time              string              `json:"time,omitempty"`
    TotalAmount       int                 `json:"total_amount"`
    Subtotal          int                 `json:"subtotal,omitempty"`
    Tax               int                 `json:"tax,omitempty"`
    Discount          int                 `json:"discount,omitempty"`
    Items             []ReceiptItem       `json:"items,omitempty"`
    PaymentMethod     string              `json:"payment_method,omitempty"`
    SuggestedCategory *SuggestedCategory  `json:"suggested_category,omitempty"`
}

type ReceiptItem struct {
    Name      string `json:"name"`
    Quantity  int    `json:"quantity"`
    UnitPrice int    `json:"unit_price"`
    Total     int    `json:"total"`
}

type SuggestedCategory struct {
    ID   uint   `json:"id"`
    Name string `json:"name"`
}
```

### DTOs

```go
// dto/receipt_dto.go

package dto

type UploadReceiptRequest struct {
    AssetID *uint `form:"asset_id"`
}

type ConfirmReceiptRequest struct {
    Amount          int    `json:"amount" binding:"required,gt=0"`
    Date            string `json:"date" binding:"required"`
    Description     string `json:"description" binding:"required"`
    CategoryID      uint   `json:"category_id" binding:"required"`
    AssetID         uint   `json:"asset_id" binding:"required"`
    TransactionType int    `json:"transaction_type" binding:"required,oneof=1 2"`
    Notes           string `json:"notes"`
}

type ReceiptScanResponse struct {
    ScanID          uint                    `json:"scan_id"`
    Status          string                  `json:"status"`
    ImageURL        string                  `json:"image_url"`
    ExtractedData   *ExtractedReceiptData   `json:"extracted_data,omitempty"`
    ConfidenceScore *float64                `json:"confidence_score,omitempty"`
    RawText         string                  `json:"raw_text,omitempty"`
    ErrorMessage    string                  `json:"error_message,omitempty"`
    CreatedAt       time.Time               `json:"created_at"`
}
```

### OCR Service Interface

```go
// services/ocr_service.go

package services

import (
    "context"
    "mime/multipart"
)

type OCRService interface {
    ExtractText(ctx context.Context, imageData []byte) (*OCRResult, error)
}

type OCRResult struct {
    RawText    string
    Confidence float64
    Blocks     []TextBlock
}

type TextBlock struct {
    Text       string
    Confidence float64
    BoundingBox BoundingBox
}

type BoundingBox struct {
    X      int
    Y      int
    Width  int
    Height int
}
```

### Google Cloud Vision Implementation

```go
// services/google_vision_ocr.go

package services

import (
    "context"
    
    vision "cloud.google.com/go/vision/apiv1"
    "cloud.google.com/go/vision/v2/apiv1/visionpb"
)

type GoogleVisionOCR struct {
    client *vision.ImageAnnotatorClient
}

func NewGoogleVisionOCR() (*GoogleVisionOCR, error) {
    ctx := context.Background()
    client, err := vision.NewImageAnnotatorClient(ctx)
    if err != nil {
        return nil, err
    }
    return &GoogleVisionOCR{client: client}, nil
}

func (g *GoogleVisionOCR) ExtractText(ctx context.Context, imageData []byte) (*OCRResult, error) {
    image := &visionpb.Image{Content: imageData}
    
    annotations, err := g.client.DetectTexts(ctx, image, nil, 50)
    if err != nil {
        return nil, err
    }
    
    if len(annotations) == 0 {
        return &OCRResult{RawText: "", Confidence: 0}, nil
    }
    
    // First annotation contains the entire text
    fullText := annotations[0].Description
    
    // Confidence from document detection
    docAnnotation, err := g.client.DetectDocumentText(ctx, image, nil)
    if err != nil {
        return &OCRResult{RawText: fullText, Confidence: 0}, nil
    }
    
    avgConfidence := 0.0
    if docAnnotation != nil && len(docAnnotation.Pages) > 0 {
        totalConf := 0.0
        count := 0
        for _, page := range docAnnotation.Pages {
            for _, block := range page.Blocks {
                totalConf += float64(block.Confidence)
                count++
            }
        }
        if count > 0 {
            avgConfidence = (totalConf / float64(count)) * 100
        }
    }
    
    return &OCRResult{
        RawText:    fullText,
        Confidence: avgConfidence,
    }, nil
}
```

### Tesseract Implementation (Self-hosted alternative)

```go
// services/tesseract_ocr.go

package services

import (
    "context"
    "os/exec"
    "strings"
)

type TesseractOCR struct {
    language string
}

func NewTesseractOCR(lang string) *TesseractOCR {
    if lang == "" {
        lang = "eng+ind" // English + Indonesian
    }
    return &TesseractOCR{language: lang}
}

func (t *TesseractOCR) ExtractText(ctx context.Context, imageData []byte) (*OCRResult, error) {
    // Write image to temp file
    tempFile, err := os.CreateTemp("", "receipt-*.jpg")
    if err != nil {
        return nil, err
    }
    defer os.Remove(tempFile.Name())
    
    if _, err := tempFile.Write(imageData); err != nil {
        return nil, err
    }
    tempFile.Close()
    
    // Run tesseract
    cmd := exec.CommandContext(ctx, "tesseract", tempFile.Name(), "stdout", "-l", t.language)
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    return &OCRResult{
        RawText:    strings.TrimSpace(string(output)),
        Confidence: 70.0, // Tesseract doesn't provide easy confidence
    }, nil
}
```

### Receipt Parser Service

```go
// services/receipt_parser_service.go

package services

import (
    "regexp"
    "strconv"
    "strings"
    "time"
    "my-api/models"
)

type ReceiptParserService interface {
    ParseReceipt(rawText string, userID uint) (*models.ExtractedReceiptData, error)
}

type receiptParserService struct {
    categoryRepo repositories.CategoryRepository
    templateRepo repositories.ReceiptTemplateRepository
}

func NewReceiptParserService(
    categoryRepo repositories.CategoryRepository,
    templateRepo repositories.ReceiptTemplateRepository,
) ReceiptParserService {
    return &receiptParserService{
        categoryRepo: categoryRepo,
        templateRepo: templateRepo,
    }
}

func (s *receiptParserService) ParseReceipt(rawText string, userID uint) (*models.ExtractedReceiptData, error) {
    lines := strings.Split(rawText, "\n")
    
    data := &models.ExtractedReceiptData{}
    
    // Extract merchant name (usually first non-empty line)
    data.MerchantName = s.extractMerchantName(lines)
    
    // Extract date
    data.Date = s.extractDate(rawText)
    
    // Extract time
    data.Time = s.extractTime(rawText)
    
    // Extract total amount
    data.TotalAmount = s.extractTotal(rawText)
    
    // Extract subtotal and tax
    data.Subtotal = s.extractSubtotal(rawText)
    data.Tax = s.extractTax(rawText)
    
    // Extract line items
    data.Items = s.extractItems(lines)
    
    // Extract payment method
    data.PaymentMethod = s.extractPaymentMethod(rawText)
    
    // Suggest category based on merchant/items
    data.SuggestedCategory = s.suggestCategory(data, userID)
    
    return data, nil
}

func (s *receiptParserService) extractMerchantName(lines []string) string {
    // Common Indonesian merchant patterns
    merchantPatterns := []string{
        `(?i)(indomaret|alfamart|alfamidi|superindo|giant|hypermart|carrefour|lotte|transmart)`,
        `(?i)(mcd|mcdonald|kfc|burger king|starbucks|jco|dunkin)`,
        `(?i)(tokopedia|shopee|gojek|grab|ovo)`,
    }
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        
        for _, pattern := range merchantPatterns {
            re := regexp.MustCompile(pattern)
            if match := re.FindString(line); match != "" {
                return strings.Title(strings.ToLower(match))
            }
        }
        
        // Return first substantial line as merchant name
        if len(line) > 3 && len(line) < 50 {
            return line
        }
    }
    
    return "Unknown Merchant"
}

func (s *receiptParserService) extractDate(text string) string {
    // Common date formats in Indonesian receipts
    datePatterns := []string{
        `(\d{2}[-/]\d{2}[-/]\d{4})`,          // DD-MM-YYYY or DD/MM/YYYY
        `(\d{4}[-/]\d{2}[-/]\d{2})`,          // YYYY-MM-DD
        `(\d{2}\s+\w{3}\s+\d{4})`,            // 11 Feb 2026
        `(\d{1,2}\s+\w+\s+\d{4})`,            // 11 February 2026
    }
    
    for _, pattern := range datePatterns {
        re := regexp.MustCompile(pattern)
        if match := re.FindString(text); match != "" {
            // Parse and normalize to YYYY-MM-DD
            return s.normalizeDate(match)
        }
    }
    
    return time.Now().Format("2006-01-02")
}

func (s *receiptParserService) normalizeDate(dateStr string) string {
    layouts := []string{
        "02-01-2006",
        "02/01/2006",
        "2006-01-02",
        "2 Jan 2006",
        "2 January 2006",
        "02 Jan 2006",
    }
    
    for _, layout := range layouts {
        if t, err := time.Parse(layout, dateStr); err == nil {
            return t.Format("2006-01-02")
        }
    }
    
    return time.Now().Format("2006-01-02")
}

func (s *receiptParserService) extractTime(text string) string {
    re := regexp.MustCompile(`(\d{2}:\d{2}(?::\d{2})?)`)
    if match := re.FindString(text); match != "" {
        return match
    }
    return ""
}

func (s *receiptParserService) extractTotal(text string) int {
    // Indonesian receipt total patterns
    totalPatterns := []string{
        `(?i)(?:total|grand total|jumlah|total bayar)[:\s]*(?:rp\.?)?[\s]*([0-9.,]+)`,
        `(?i)(?:rp\.?)?[\s]*([0-9.,]+)[\s]*(?:total|grand total)`,
        `(?i)tunai[:\s]*(?:rp\.?)?[\s]*([0-9.,]+)`,
    }
    
    for _, pattern := range totalPatterns {
        re := regexp.MustCompile(pattern)
        if matches := re.FindStringSubmatch(text); len(matches) > 1 {
            return s.parseAmount(matches[1])
        }
    }
    
    return 0
}

func (s *receiptParserService) extractSubtotal(text string) int {
    re := regexp.MustCompile(`(?i)(?:sub\s*total|subtotal)[:\s]*(?:rp\.?)?[\s]*([0-9.,]+)`)
    if matches := re.FindStringSubmatch(text); len(matches) > 1 {
        return s.parseAmount(matches[1])
    }
    return 0
}

func (s *receiptParserService) extractTax(text string) int {
    re := regexp.MustCompile(`(?i)(?:tax|pajak|ppn|vat)[:\s]*(?:rp\.?)?[\s]*([0-9.,]+)`)
    if matches := re.FindStringSubmatch(text); len(matches) > 1 {
        return s.parseAmount(matches[1])
    }
    return 0
}

func (s *receiptParserService) parseAmount(amountStr string) int {
    // Remove common separators (dots, commas, spaces)
    cleaned := strings.ReplaceAll(amountStr, ".", "")
    cleaned = strings.ReplaceAll(cleaned, ",", "")
    cleaned = strings.ReplaceAll(cleaned, " ", "")
    
    amount, _ := strconv.Atoi(cleaned)
    return amount
}

func (s *receiptParserService) extractItems(lines []string) []models.ReceiptItem {
    items := []models.ReceiptItem{}
    
    // Pattern: item name followed by quantity and price
    itemPattern := regexp.MustCompile(`^(.+?)\s+(\d+)\s*[xX@]\s*([0-9.,]+)\s*=?\s*([0-9.,]+)?`)
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if matches := itemPattern.FindStringSubmatch(line); len(matches) >= 4 {
            qty, _ := strconv.Atoi(matches[2])
            unitPrice := s.parseAmount(matches[3])
            total := unitPrice * qty
            
            if len(matches) > 4 && matches[4] != "" {
                total = s.parseAmount(matches[4])
            }
            
            items = append(items, models.ReceiptItem{
                Name:      strings.TrimSpace(matches[1]),
                Quantity:  qty,
                UnitPrice: unitPrice,
                Total:     total,
            })
        }
    }
    
    return items
}

func (s *receiptParserService) extractPaymentMethod(text string) string {
    methods := map[string][]string{
        "Cash":        {"tunai", "cash", "uang tunai"},
        "Debit Card":  {"debit", "debet", "kartu debit"},
        "Credit Card": {"credit", "kredit", "kartu kredit"},
        "QRIS":        {"qris", "qr code"},
        "E-Wallet":    {"gopay", "ovo", "dana", "shopeepay", "linkaja"},
    }
    
    textLower := strings.ToLower(text)
    
    for method, keywords := range methods {
        for _, keyword := range keywords {
            if strings.Contains(textLower, keyword) {
                return method
            }
        }
    }
    
    return ""
}

func (s *receiptParserService) suggestCategory(data *models.ExtractedReceiptData, userID uint) *models.SuggestedCategory {
    merchantLower := strings.ToLower(data.MerchantName)
    
    // Merchant to category mapping
    categoryMapping := map[string]string{
        "indomaret":  "Groceries",
        "alfamart":   "Groceries",
        "superindo":  "Groceries",
        "mcdonald":   "Food & Dining",
        "kfc":        "Food & Dining",
        "starbucks":  "Food & Dining",
        "tokopedia":  "Shopping",
        "shopee":     "Shopping",
        "gojek":      "Transportation",
        "grab":       "Transportation",
        "pertamina":  "Transportation",
        "shell":      "Transportation",
    }
    
    for keyword, categoryName := range categoryMapping {
        if strings.Contains(merchantLower, keyword) {
            // Find category by name for this user
            categories, _ := s.categoryRepo.GetByUserID(userID)
            for _, cat := range categories {
                if strings.EqualFold(cat.CategoryName, categoryName) {
                    return &models.SuggestedCategory{
                        ID:   cat.ID,
                        Name: cat.CategoryName,
                    }
                }
            }
        }
    }
    
    return nil
}
```

### Receipt Controller

```go
// controllers/receipt_controller.go

package controllers

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "my-api/dto"
    "my-api/services"
    "my-api/utils"
)

type ReceiptController struct {
    receiptService services.ReceiptService
}

func NewReceiptController(receiptService services.ReceiptService) *ReceiptController {
    return &ReceiptController{receiptService: receiptService}
}

func (c *ReceiptController) UploadReceipt(ctx *gin.Context) {
    userID := ctx.GetUint("userID")
    
    file, header, err := ctx.Request.FormFile("image")
    if err != nil {
        utils.ErrorResponse(ctx, http.StatusBadRequest, "Image file is required")
        return
    }
    defer file.Close()
    
    // Validate file type
    contentType := header.Header.Get("Content-Type")
    if !isValidImageType(contentType) {
        utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid image format. Supported: JPEG, PNG")
        return
    }
    
    // Validate file size (max 10MB)
    if header.Size > 10*1024*1024 {
        utils.ErrorResponse(ctx, http.StatusBadRequest, "Image size must not exceed 10MB")
        return
    }
    
    var req dto.UploadReceiptRequest
    ctx.ShouldBind(&req)
    
    result, err := c.receiptService.ProcessReceiptUpload(userID, file, header, req.AssetID)
    if err != nil {
        utils.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
        return
    }
    
    utils.SuccessResponse(ctx, http.StatusAccepted, "Receipt uploaded and processing started", result)
}

func (c *ReceiptController) GetScanStatus(ctx *gin.Context) {
    userID := ctx.GetUint("userID")
    scanID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
    
    result, err := c.receiptService.GetScanResult(uint(scanID), userID)
    if err != nil {
        utils.ErrorResponse(ctx, http.StatusNotFound, "Receipt scan not found")
        return
    }
    
    utils.SuccessResponse(ctx, http.StatusOK, "Receipt scan retrieved", result)
}

func (c *ReceiptController) ConfirmAndCreateTransaction(ctx *gin.Context) {
    userID := ctx.GetUint("userID")
    scanID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
    
    var req dto.ConfirmReceiptRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
        return
    }
    
    result, err := c.receiptService.ConfirmAndCreateTransaction(uint(scanID), userID, &req)
    if err != nil {
        utils.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
        return
    }
    
    utils.SuccessResponse(ctx, http.StatusCreated, "Transaction created from receipt", result)
}

func (c *ReceiptController) ListScans(ctx *gin.Context) {
    userID := ctx.GetUint("userID")
    status := ctx.Query("status")
    page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
    
    result, total, err := c.receiptService.ListUserScans(userID, status, page, limit)
    if err != nil {
        utils.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
        return
    }
    
    utils.PaginatedResponse(ctx, http.StatusOK, "Scans retrieved", result, page, limit, total)
}

func (c *ReceiptController) DeleteScan(ctx *gin.Context) {
    userID := ctx.GetUint("userID")
    scanID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
    
    if err := c.receiptService.DeleteScan(uint(scanID), userID); err != nil {
        utils.ErrorResponse(ctx, http.StatusNotFound, "Scan not found")
        return
    }
    
    utils.SuccessResponse(ctx, http.StatusOK, "Scan deleted", nil)
}

func isValidImageType(contentType string) bool {
    validTypes := map[string]bool{
        "image/jpeg": true,
        "image/jpg":  true,
        "image/png":  true,
        "image/webp": true,
    }
    return validTypes[contentType]
}
```

### Routes

```go
// routes/routes.go (add to SetupRoutes)

// Receipt OCR routes
receiptController := controllers.NewReceiptController(receiptService)

receipts := v2.Group("/receipts")
receipts.Use(middleware.AuthMiddleware())
{
    receipts.POST("/scan", receiptController.UploadReceipt)
    receipts.GET("/scan/:id", receiptController.GetScanStatus)
    receipts.POST("/scan/:id/confirm", receiptController.ConfirmAndCreateTransaction)
    receipts.GET("", receiptController.ListScans)
    receipts.DELETE("/scan/:id", receiptController.DeleteScan)
}
```

---

## Image Storage

### Option 1: Local Storage

```go
// services/local_storage.go

func (s *localStorage) UploadImage(userID uint, file multipart.File, filename string) (string, error) {
    uploadDir := fmt.Sprintf("uploads/receipts/%d", userID)
    os.MkdirAll(uploadDir, 0755)
    
    uniqueName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filename)
    filePath := filepath.Join(uploadDir, uniqueName)
    
    dst, err := os.Create(filePath)
    if err != nil {
        return "", err
    }
    defer dst.Close()
    
    if _, err := io.Copy(dst, file); err != nil {
        return "", err
    }
    
    return "/" + filePath, nil
}
```

### Option 2: AWS S3

```go
// services/s3_storage.go

func (s *s3Storage) UploadImage(userID uint, file multipart.File, filename string, contentType string) (string, error) {
    key := fmt.Sprintf("receipts/%d/%d_%s", userID, time.Now().UnixNano(), filename)
    
    _, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
        Bucket:      aws.String(s.bucket),
        Key:         aws.String(key),
        Body:        file,
        ContentType: aws.String(contentType),
    })
    if err != nil {
        return "", err
    }
    
    return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key), nil
}
```

---

## Dependencies

```go
// go.mod additions

require (
    cloud.google.com/go/vision v1.2.0       // Google Cloud Vision
    github.com/aws/aws-sdk-go-v2 v1.24.0    // AWS S3 (optional)
)
```

---

## Environment Variables

```env
# OCR Provider (google_vision, tesseract, aws_textract)
OCR_PROVIDER=google_vision

# Google Cloud Vision
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json

# AWS (if using S3 or Textract)
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
AWS_REGION=ap-southeast-1
S3_BUCKET=money-manage-receipts

# Storage (local, s3)
STORAGE_PROVIDER=local
LOCAL_UPLOAD_PATH=./uploads
```

---

## Performance Considerations

1. **Async Processing**: OCR runs in background goroutine
2. **Image Optimization**: Resize large images before OCR
3. **Caching**: Cache frequently seen merchant templates
4. **Rate Limiting**: Limit uploads per user (e.g., 50/day)
5. **Queue System**: For high traffic, use Redis/RabbitMQ queue

---

## Error Handling

| Error Code | Description | Action |
|------------|-------------|--------|
| `INVALID_IMAGE` | Unsupported format/corrupt | Return immediately |
| `IMAGE_TOO_LARGE` | > 10MB | Reject upload |
| `OCR_FAILED` | OCR service error | Retry up to 3 times |
| `PARSE_FAILED` | Cannot extract data | Return raw text for manual entry |
| `LOW_CONFIDENCE` | < 50% confidence | Warn user, allow manual correction |
