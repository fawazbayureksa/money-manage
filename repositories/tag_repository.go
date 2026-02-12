package repositories

import (
	"errors"
	"my-api/dto"
	"my-api/models"
	"time"

	"gorm.io/gorm"
)

// TagRepository defines the interface for tag data operations
type TagRepository interface {
	Create(tag *models.Tag) error
	FindByID(id, userID uint) (*models.Tag, error)
	FindByUserID(userID uint, sortBy string) ([]models.Tag, error)
	Update(tag *models.Tag) error
	Delete(id, userID uint) error
	FindByName(userID uint, name string) (*models.Tag, error)
	IncrementUsage(tagID uint) error
	GetTagsByCategory(userID, categoryID uint, days int) ([]CategoryTagUsage, error)
	GetSpendingByTag(userID uint, startDate, endDate time.Time) ([]dto.TagSpending, error)
}

type tagRepository struct {
	db *gorm.DB
}

// NewTagRepository creates a new tag repository
func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

// Create creates a new tag
func (r *tagRepository) Create(tag *models.Tag) error {
	return r.db.Create(tag).Error
}

// FindByID finds a tag by ID and user ID
func (r *tagRepository) FindByID(id, userID uint) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&tag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tag not found")
		}
		return nil, err
	}
	return &tag, nil
}

// FindByUserID finds all tags for a user
func (r *tagRepository) FindByUserID(userID uint, sortBy string) ([]models.Tag, error) {
	var tags []models.Tag
	query := r.db.Where("user_id = ?", userID)

	// Sort by usage count (most used first) or by name
	if sortBy == "usage" {
		query = query.Order("usage_count DESC, name ASC")
	} else {
		query = query.Order("name ASC")
	}

	err := query.Find(&tags).Error
	return tags, err
}

// Update updates a tag
func (r *tagRepository) Update(tag *models.Tag) error {
	return r.db.Model(tag).Where("id = ? AND user_id = ?", tag.ID, tag.UserID).
		Updates(map[string]interface{}{
			"name":  tag.Name,
			"color": tag.Color,
			"icon":  tag.Icon,
		}).Error
}

// Delete soft deletes a tag
func (r *tagRepository) Delete(id, userID uint) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Tag{})
	if result.RowsAffected == 0 {
		return errors.New("tag not found")
	}
	return result.Error
}

// FindByName finds a tag by name for a user
func (r *tagRepository) FindByName(userID uint, name string) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.Where("user_id = ? AND name = ?", userID, name).First(&tag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

// IncrementUsage increments the usage count of a tag
func (r *tagRepository) IncrementUsage(tagID uint) error {
	return r.db.Model(&models.Tag{}).
		Where("id = ?", tagID).
		UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1)).Error
}

// CategoryTagUsage represents tag usage statistics for a category
type CategoryTagUsage struct {
	TagID uint
	Count int
}

// GetTagsByCategory gets tags frequently used with a specific category
func (r *tagRepository) GetTagsByCategory(userID, categoryID uint, days int) ([]CategoryTagUsage, error) {
	var results []CategoryTagUsage
	cutoffDate := time.Now().AddDate(0, 0, -days)

	err := r.db.Raw(`
		SELECT tt.tag_id as tag_id, COUNT(*) as count
		FROM transaction_tags tt
		JOIN transactions t ON tt.transaction_id = t.id
		JOIN tags tg ON tt.tag_id = tg.id
		WHERE tg.user_id = ?
		  AND t.category_id = ?
		  AND t.date >= ?
		GROUP BY tt.tag_id
		ORDER BY count DESC
		LIMIT 10
	`, userID, categoryID, cutoffDate).Scan(&results).Error

	return results, err
}

// GetSpendingByTag gets spending analytics grouped by tag
func (r *tagRepository) GetSpendingByTag(userID uint, startDate, endDate time.Time) ([]dto.TagSpending, error) {
	var results []dto.TagSpending

	err := r.db.Raw(`
		SELECT 
			t.id,
			t.name,
			t.color,
			t.icon,
			t.usage_count,
			SUM(tx.amount) as total_amount,
			COUNT(tx.id) as transaction_count,
			AVG(tx.amount) as avg_amount
		FROM tags t
		JOIN transaction_tags tt ON t.id = tt.tag_id
		JOIN transactions tx ON tt.transaction_id = tx.id
		WHERE t.user_id = ?
		  AND tx.transaction_type = 2
		  AND tx.date BETWEEN ? AND ?
		GROUP BY t.id, t.name, t.color, t.icon, t.usage_count
		ORDER BY total_amount DESC
	`, userID, startDate, endDate).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Build the Tag objects and calculate averages
	for i := range results {
		results[i].Tag = models.Tag{
			ID:         results[i].Tag.ID,
			Name:       results[i].Tag.Name,
			Color:      results[i].Tag.Color,
			Icon:       results[i].Tag.Icon,
			UsageCount: results[i].Tag.UsageCount,
		}
		if results[i].TransactionCount > 0 {
			results[i].AvgAmount = results[i].TotalAmount / float64(results[i].TransactionCount)
		}
	}

	return results, nil
}
