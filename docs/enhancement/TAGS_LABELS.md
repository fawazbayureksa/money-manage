# Tags & Labels Implementation Plan

## Problem Statement

Categories alone don't capture the full context of transactions:
- Same restaurant visit could be "date night", "business lunch", or "family dinner"
- Shopping trip for "birthday gift" vs "personal"
- Travel expenses for "vacaction" vs "business trip"
- Project-based spending across multiple categories

Users need a flexible way to add custom context beyond categories.

---

## Solution Overview

User-defined tags that can be attached to any transaction, enabling:
- Multi-tag support per transaction
- Custom tag colors
- Tag-based filtering and analytics
- Auto-suggest based on history

---

## Database Schema

### Table: `tags`

```sql
CREATE TABLE tags (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) DEFAULT '#6366F1' COMMENT 'Hex color code',
    icon VARCHAR(10) NULL COMMENT 'Emoji',
    usage_count INT DEFAULT 0 COMMENT 'For sorting by popularity',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    UNIQUE KEY unique_user_tag (user_id, name),
    INDEX idx_user_usage (user_id, usage_count DESC),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### Table: `transaction_tags`

```sql
CREATE TABLE transaction_tags (
    transaction_id BIGINT UNSIGNED NOT NULL,
    tag_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (transaction_id, tag_id),
    FOREIGN KEY (transaction_id) REFERENCES transactions_v2(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
```

---

## API Endpoints

### Tags Management

#### List User Tags
```http
GET /api/v2/tags?sort=usage
Authorization: Bearer {token}
```

#### Response
```json
{
    "success": true,
    "data": [
        {"id": 1, "name": "Date Night", "color": "#EC4899", "icon": "❤️", "usage_count": 12},
        {"id": 2, "name": "Business", "color": "#3B82F6", "icon": "💼", "usage_count": 8},
        {"id": 3, "name": "Family", "color": "#10B981", "icon": "👨‍👩‍👧", "usage_count": 5},
        {"id": 4, "name": "Birthday", "color": "#F59E0B", "icon": "🎂", "usage_count": 2}
    ]
}
```

#### Create Tag
```http
POST /api/v2/tags
Authorization: Bearer {token}

{
    "name": "Vacation 2026",
    "color": "#F97316",
    "icon": "🏖️"
}
```

#### Update Tag
```http
PUT /api/v2/tags/{id}
Authorization: Bearer {token}

{
    "name": "Beach Vacation 2026",
    "color": "#06B6D4"
}
```

#### Delete Tag
```http
DELETE /api/v2/tags/{id}
Authorization: Bearer {token}
```

### Transaction with Tags

#### Create Transaction with Tags
```http
POST /api/v2/transactions
Authorization: Bearer {token}

{
    "description": "Dinner at Sushi Tei",
    "amount": 450000,
    "transaction_type": "Expense",
    "category_id": 3,
    "asset_id": 1,
    "date": "2026-02-11",
    "tag_ids": [1, 3]
}
```

#### Add Tags to Existing Transaction
```http
POST /api/v2/transactions/{id}/tags
Authorization: Bearer {token}

{
    "tag_ids": [1, 3]
}
```

#### Remove Tag from Transaction
```http
DELETE /api/v2/transactions/{id}/tags/{tag_id}
Authorization: Bearer {token}
```

### Tag-based Queries

#### Get Transactions by Tag
```http
GET /api/v2/transactions?tag_id=1&start_date=2026-01-01&end_date=2026-02-28
Authorization: Bearer {token}
```

#### Get Spending by Tag
```http
GET /api/v2/analytics/spending-by-tag?start_date=2026-01-01&end_date=2026-02-28
Authorization: Bearer {token}
```

#### Response
```json
{
    "success": true,
    "data": [
        {
            "tag": {"id": 1, "name": "Date Night", "color": "#EC4899", "icon": "❤️"},
            "total_amount": 2500000,
            "transaction_count": 12,
            "avg_amount": 208333
        },
        {
            "tag": {"id": 2, "name": "Business", "color": "#3B82F6", "icon": "💼"},
            "total_amount": 1800000,
            "transaction_count": 8,
            "avg_amount": 225000
        }
    ],
    "period": {
        "start_date": "2026-01-01",
        "end_date": "2026-02-28"
    }
}
```

#### Tag Suggestions (based on category and description)
```http
GET /api/v2/tags/suggest?category_id=3&description=dinner
Authorization: Bearer {token}
```

#### Response
```json
{
    "success": true,
    "data": [
        {"id": 1, "name": "Date Night", "confidence": 0.8},
        {"id": 3, "name": "Family", "confidence": 0.6},
        {"id": 2, "name": "Business", "confidence": 0.4}
    ]
}
```

---

## Go Implementation

### Model

```go
// models/tag.go

package models

import (
    "my-api/utils"
    "gorm.io/gorm"
)

type Tag struct {
    ID         uint            `gorm:"primaryKey" json:"id"`
    UserID     uint            `gorm:"not null;index" json:"user_id"`
    Name       string          `gorm:"size:50;not null" json:"name"`
    Color      string          `gorm:"size:7;default:'#6366F1'" json:"color"`
    Icon       string          `gorm:"size:10" json:"icon,omitempty"`
    UsageCount int             `gorm:"default:0" json:"usage_count"`
    
    CreatedAt  utils.CustomTime `json:"created_at"`
    UpdatedAt  utils.CustomTime `json:"updated_at"`
    DeletedAt  gorm.DeletedAt  `gorm:"index" json:"-"`
}

type TransactionTag struct {
    TransactionID uint `gorm:"primaryKey" json:"transaction_id"`
    TagID         uint `gorm:"primaryKey" json:"tag_id"`
    
    Tag Tag `gorm:"foreignKey:TagID" json:"tag,omitempty"`
}

// Update TransactionV2 model
type TransactionV2 struct {
    // ... existing fields ...
    
    Tags []Tag `gorm:"many2many:transaction_tags;" json:"tags,omitempty"`
}
```

### Service

```go
// services/tag_service.go

package services

import (
    "errors"
    "my-api/models"
    "my-api/repositories"
)

type TagService interface {
    CreateTag(userID uint, req *CreateTagRequest) (*models.Tag, error)
    GetTags(userID uint, sortBy string) ([]models.Tag, error)
    UpdateTag(id, userID uint, req *UpdateTagRequest) (*models.Tag, error)
    DeleteTag(id, userID uint) error
    AddTagsToTransaction(transactionID, userID uint, tagIDs []uint) error
    RemoveTagFromTransaction(transactionID, userID uint, tagID uint) error
    SuggestTags(userID uint, categoryID uint, description string) ([]TagSuggestion, error)
    GetSpendingByTag(userID uint, startDate, endDate time.Time) ([]TagSpending, error)
}

type tagService struct {
    repo            repositories.TagRepository
    transactionRepo repositories.TransactionV2Repository
}

func (s *tagService) SuggestTags(userID uint, categoryID uint, description string) ([]TagSuggestion, error) {
    // Get all user tags
    tags, _ := s.repo.FindByUserID(userID)
    
    // Get historical tag usage for this category
    categoryTags, _ := s.repo.GetTagsByCategory(userID, categoryID, 30) // Last 30 days
    
    // Simple scoring based on:
    // 1. Previously used with this category
    // 2. Keyword match in description
    // 3. Overall usage frequency
    
    suggestions := make([]TagSuggestion, 0)
    keywords := strings.Fields(strings.ToLower(description))
    
    for _, tag := range tags {
        score := 0.0
        
        // Category match
        for _, ct := range categoryTags {
            if ct.TagID == tag.ID {
                score += 0.5 * float64(ct.Count) / float64(len(categoryTags))
            }
        }
        
        // Keyword match
        tagLower := strings.ToLower(tag.Name)
        for _, kw := range keywords {
            if strings.Contains(tagLower, kw) || strings.Contains(kw, tagLower) {
                score += 0.3
                break
            }
        }
        
        // Usage frequency
        score += 0.2 * float64(tag.UsageCount) / 100.0
        
        if score > 0.1 {
            suggestions = append(suggestions, TagSuggestion{
                Tag:        tag,
                Confidence: score,
            })
        }
    }
    
    // Sort by confidence
    sort.Slice(suggestions, func(i, j int) bool {
        return suggestions[i].Confidence > suggestions[j].Confidence
    })
    
    // Return top 5
    if len(suggestions) > 5 {
        suggestions = suggestions[:5]
    }
    
    return suggestions, nil
}

func (s *tagService) GetSpendingByTag(userID uint, startDate, endDate time.Time) ([]TagSpending, error) {
    return s.repo.GetSpendingByTag(userID, startDate, endDate)
}
```

### Repository

```go
// repositories/tag_repository.go

func (r *tagRepository) GetSpendingByTag(userID uint, startDate, endDate time.Time) ([]TagSpending, error) {
    var results []TagSpending
    
    err := r.db.Raw(`
        SELECT 
            t.id as tag_id,
            t.name as tag_name,
            t.color,
            t.icon,
            SUM(tx.amount) as total_amount,
            COUNT(tx.id) as transaction_count
        FROM tags t
        JOIN transaction_tags tt ON t.id = tt.tag_id
        JOIN transactions_v2 tx ON tt.transaction_id = tx.id
        WHERE t.user_id = ?
          AND tx.transaction_type = 2
          AND tx.date BETWEEN ? AND ?
        GROUP BY t.id, t.name, t.color, t.icon
        ORDER BY total_amount DESC
    `, userID, startDate, endDate).Scan(&results).Error
    
    return results, err
}

func (r *tagRepository) IncrementUsage(tagID uint) error {
    return r.db.Model(&models.Tag{}).
        Where("id = ?", tagID).
        UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}
```

---

## Frontend Integration Guide

### Tag Selection UI

```
┌─────────────────────────────────────────┐
│  Add Tags                           ✕   │
├─────────────────────────────────────────┤
│                                         │
│  🔍 Search or create tag...             │
│  ┌─────────────────────────────────┐   │
│  │                                 │   │
│  └─────────────────────────────────┘   │
│                                         │
│  SUGGESTED                              │
│  ┌───────────┐ ┌───────────┐           │
│  │ ❤️ Date   │ │ 👨‍👩‍👧 Family│           │
│  │   Night   │ │           │           │
│  └───────────┘ └───────────┘           │
│                                         │
│  RECENT                                 │
│  ┌───────────┐ ┌───────────┐ ┌───────┐ │
│  │ 💼 Busi-  │ │ 🎂 Birth- │ │ + New │ │
│  │   ness    │ │   day     │ │       │ │
│  └───────────┘ └───────────┘ └───────┘ │
│                                         │
│  ALL TAGS                               │
│  ┌───────────┐ ┌───────────┐           │
│  │ 🏖️ Vaca-  │ │ 🎁 Gift   │           │
│  │   tion    │ │           │           │
│  └───────────┘ └───────────┘           │
│                                         │
│  ─────────────────────────────────────  │
│  Selected: ❤️ Date Night, 👨‍👩‍👧 Family    │
│                                         │
│  [ Done ]                               │
└─────────────────────────────────────────┘
```

### Transaction with Tags

```
┌─────────────────────────────────────────┐
│  ↓ Feb 11, 2026                         │
├─────────────────────────────────────────┤
│  🍜 Dinner at Sushi Tei                 │
│  -Rp 450,000                            │
│  🍔 Food & Dining                       │
│                                         │
│  ┌───────────┐ ┌───────────┐           │
│  │ ❤️ Date   │ │ 👨‍👩‍👧 Family│           │
│  │   Night   │ │           │           │
│  └───────────┘ └───────────┘           │
└─────────────────────────────────────────┘
```

### Spending by Tag Analytics

```
┌─────────────────────────────────────────┐
│  ← Spending by Tags                     │
├─────────────────────────────────────────┤
│                                         │
│  Jan 1 - Feb 28, 2026                   │
│                                         │
│  ❤️ Date Night                          │
│  ████████████████████░░░░ Rp 2,500,000  │
│  12 transactions • Avg Rp 208,333       │
│                                         │
│  💼 Business                            │
│  ██████████████░░░░░░░░░░ Rp 1,800,000  │
│  8 transactions • Avg Rp 225,000        │
│                                         │
│  👨‍👩‍👧 Family                              │
│  ██████████░░░░░░░░░░░░░░ Rp 1,200,000  │
│  5 transactions • Avg Rp 240,000        │
│                                         │
│  🎂 Birthday                            │
│  █████░░░░░░░░░░░░░░░░░░░ Rp 500,000    │
│  2 transactions • Avg Rp 250,000        │
│                                         │
└─────────────────────────────────────────┘
```

### Tag Management Screen

```
┌─────────────────────────────────────────┐
│  ← Manage Tags                          │
├─────────────────────────────────────────┤
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ ❤️ Date Night                   │   │
│  │ #EC4899 • Used 12 times      ⋮  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 💼 Business                     │   │
│  │ #3B82F6 • Used 8 times       ⋮  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 👨‍👩‍👧 Family                       │   │
│  │ #10B981 • Used 5 times       ⋮  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 🎂 Birthday                     │   │
│  │ #F59E0B • Used 2 times       ⋮  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  [ + Create New Tag ]                   │
└─────────────────────────────────────────┘
```

---

## Migration Files

```sql
-- db/migrations/YYYYMMDDHHMMSS_create_tags.up.sql

CREATE TABLE tags (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) DEFAULT '#6366F1',
    icon VARCHAR(10) NULL,
    usage_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    UNIQUE KEY unique_user_tag (user_id, name),
    INDEX idx_user_usage (user_id, usage_count DESC),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE transaction_tags (
    transaction_id BIGINT UNSIGNED NOT NULL,
    tag_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (transaction_id, tag_id),
    FOREIGN KEY (transaction_id) REFERENCES transactions_v2(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
```

---

## Use Cases

### 1. Date Night Tracking
- Tag all romantic outings as "Date Night"
- View total spent on dates per month
- Compare with budget or savings goals

### 2. Business Expense Reporting
- Tag work-related expenses
- Generate reports for reimbursement
- Track tax-deductible expenses

### 3. Event/Trip Tracking
- Create tag "Bali Trip 2026"
- Tag all related expenses (flights, hotels, food)
- See total trip cost across categories

### 4. Gift Tracking
- Tag purchases as gifts
- Track gift spending per person or occasion
- Plan better for holidays
