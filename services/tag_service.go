package services

import (
	"errors"
	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
	"sort"
	"strings"
	"time"
)

// TagService defines the interface for tag business logic
type TagService interface {
	CreateTag(userID uint, req *dto.CreateTagRequest) (*models.Tag, error)
	GetTags(userID uint, sortBy string) ([]models.Tag, error)
	UpdateTag(id, userID uint, req *dto.UpdateTagRequest) (*models.Tag, error)
	DeleteTag(id, userID uint) error
	GetTagByID(id, userID uint) (*models.Tag, error)
	SuggestTags(userID, categoryID uint, description string) ([]dto.TagSuggestion, error)
	GetSpendingByTag(userID uint, startDate, endDate time.Time) (*dto.TagSpendingResponse, error)
}

type tagService struct {
	repo repositories.TagRepository
}

// NewTagService creates a new tag service
func NewTagService(repo repositories.TagRepository) TagService {
	return &tagService{repo: repo}
}

// CreateTag creates a new tag for a user
func (s *tagService) CreateTag(userID uint, req *dto.CreateTagRequest) (*models.Tag, error) {
	// Check if tag with same name already exists
	existing, err := s.repo.FindByName(userID, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("tag with this name already exists")
	}

	// Set default color if not provided
	color := req.Color
	if color == "" {
		color = "#6366F1"
	}

	tag := &models.Tag{
		UserID: userID,
		Name:   req.Name,
		Color:  color,
		Icon:   req.Icon,
	}

	err = s.repo.Create(tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// GetTags gets all tags for a user
func (s *tagService) GetTags(userID uint, sortBy string) ([]models.Tag, error) {
	return s.repo.FindByUserID(userID, sortBy)
}

// GetTagByID gets a tag by ID
func (s *tagService) GetTagByID(id, userID uint) (*models.Tag, error) {
	return s.repo.FindByID(id, userID)
}

// UpdateTag updates a tag
func (s *tagService) UpdateTag(id, userID uint, req *dto.UpdateTagRequest) (*models.Tag, error) {
	// Get existing tag
	tag, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		// Check if new name conflicts with existing tags
		existing, err := s.repo.FindByName(userID, *req.Name)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, errors.New("tag with this name already exists")
		}
		tag.Name = *req.Name
	}
	if req.Color != nil {
		tag.Color = *req.Color
	}
	if req.Icon != nil {
		tag.Icon = *req.Icon
	}

	err = s.repo.Update(tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// DeleteTag deletes a tag
func (s *tagService) DeleteTag(id, userID uint) error {
	return s.repo.Delete(id, userID)
}

// SuggestTags suggests tags based on category and description
func (s *tagService) SuggestTags(userID, categoryID uint, description string) ([]dto.TagSuggestion, error) {
	// Get all user tags
	tags, err := s.repo.FindByUserID(userID, "usage")
	if err != nil {
		return nil, err
	}

	if len(tags) == 0 {
		return []dto.TagSuggestion{}, nil
	}

	// Get historical tag usage for this category
	categoryTags, err := s.repo.GetTagsByCategory(userID, categoryID, 30) // Last 30 days
	if err != nil {
		categoryTags = []repositories.CategoryTagUsage{}
	}

	// Create a map for quick lookup of category tag counts
	categoryTagMap := make(map[uint]int)
	maxCategoryCount := 0
	for _, ct := range categoryTags {
		categoryTagMap[ct.TagID] = ct.Count
		if ct.Count > maxCategoryCount {
			maxCategoryCount = ct.Count
		}
	}

	// Find max usage count for normalization
	maxUsage := 0
	for _, tag := range tags {
		if tag.UsageCount > maxUsage {
			maxUsage = tag.UsageCount
		}
	}

	// Simple scoring based on:
	// 1. Previously used with this category (50% weight)
	// 2. Keyword match in description (30% weight)
	// 3. Overall usage frequency (20% weight)
	suggestions := make([]dto.TagSuggestion, 0)
	keywords := strings.Fields(strings.ToLower(description))

	for _, tag := range tags {
		score := 0.0

		// Category match
		if categoryCount, exists := categoryTagMap[tag.ID]; exists && maxCategoryCount > 0 {
			score += 0.5 * float64(categoryCount) / float64(maxCategoryCount)
		}

		// Keyword match
		tagLower := strings.ToLower(tag.Name)
		for _, kw := range keywords {
			if len(kw) < 3 {
				continue // Skip short keywords
			}
			if strings.Contains(tagLower, kw) || strings.Contains(kw, tagLower) {
				score += 0.3
				break
			}
		}

		// Usage frequency
		if maxUsage > 0 {
			score += 0.2 * float64(tag.UsageCount) / float64(maxUsage)
		}

		if score > 0.1 {
			suggestions = append(suggestions, dto.TagSuggestion{
				ID:         tag.ID,
				Name:       tag.Name,
				Confidence: score,
			})
		}
	}

	// Sort by confidence (descending) using standard library
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Confidence > suggestions[j].Confidence
	})

	// Return top 5
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}

	return suggestions, nil
}

// GetSpendingByTag gets spending analytics grouped by tag
func (s *tagService) GetSpendingByTag(userID uint, startDate, endDate time.Time) (*dto.TagSpendingResponse, error) {
	spendingData, err := s.repo.GetSpendingByTag(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	response := &dto.TagSpendingResponse{
		Data: spendingData,
		Period: dto.PeriodInfo{
			StartDate: startDate.Format("2006-01-02"),
			EndDate:   endDate.Format("2006-01-02"),
		},
	}

	return response, nil
}
