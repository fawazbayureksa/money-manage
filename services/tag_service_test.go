package services

import (
	"errors"
	"testing"
	"time"

	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
)

// --- Mock TagRepository ---

type mockTagRepository struct {
	tags        map[uint]*models.Tag
	nextID      uint
	createError error
	deleteError error
	updateError error
}

func newMockTagRepo() *mockTagRepository {
	return &mockTagRepository{
		tags:   make(map[uint]*models.Tag),
		nextID: 1,
	}
}

func (m *mockTagRepository) Create(tag *models.Tag) error {
	if m.createError != nil {
		return m.createError
	}
	tag.ID = m.nextID
	m.nextID++
	copy := *tag
	m.tags[copy.ID] = &copy
	return nil
}

func (m *mockTagRepository) FindByID(id, userID uint) (*models.Tag, error) {
	if t, ok := m.tags[id]; ok && t.UserID == userID {
		return t, nil
	}
	return nil, errors.New("tag not found")
}

func (m *mockTagRepository) FindByUserID(userID uint, sortBy string) ([]models.Tag, error) {
	var result []models.Tag
	for _, t := range m.tags {
		if t.UserID == userID {
			result = append(result, *t)
		}
	}
	return result, nil
}

func (m *mockTagRepository) Update(tag *models.Tag) error {
	if m.updateError != nil {
		return m.updateError
	}
	copy := *tag
	m.tags[copy.ID] = &copy
	return nil
}

func (m *mockTagRepository) Delete(id, userID uint) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	if _, ok := m.tags[id]; !ok {
		return errors.New("tag not found")
	}
	delete(m.tags, id)
	return nil
}

func (m *mockTagRepository) FindByName(userID uint, name string) (*models.Tag, error) {
	for _, t := range m.tags {
		if t.UserID == userID && t.Name == name {
			return t, nil
		}
	}
	return nil, nil
}

func (m *mockTagRepository) IncrementUsage(tagID uint) error {
	if t, ok := m.tags[tagID]; ok {
		t.UsageCount++
	}
	return nil
}

func (m *mockTagRepository) GetTagsByCategory(userID, categoryID uint, days int) ([]repositories.CategoryTagUsage, error) {
	return []repositories.CategoryTagUsage{}, nil
}

func (m *mockTagRepository) GetSpendingByTag(userID uint, startDate, endDate time.Time) ([]dto.TagSpending, error) {
	return []dto.TagSpending{}, nil
}

// --- Tests ---

func TestTagService_CreateTag(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	tag, err := svc.CreateTag(1, &dto.CreateTagRequest{
		Name:  "Groceries",
		Color: "#FF5733",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if tag.Name != "Groceries" {
		t.Errorf("Expected name 'Groceries', got '%s'", tag.Name)
	}
	if tag.Color != "#FF5733" {
		t.Errorf("Expected color '#FF5733', got '%s'", tag.Color)
	}
}

func TestTagService_CreateTag_DefaultColor(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	tag, err := svc.CreateTag(1, &dto.CreateTagRequest{Name: "Travel"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if tag.Color != "#6366F1" {
		t.Errorf("Expected default color '#6366F1', got '%s'", tag.Color)
	}
}

func TestTagService_CreateTag_DuplicateName(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	svc.CreateTag(1, &dto.CreateTagRequest{Name: "Duplicate"})
	_, err := svc.CreateTag(1, &dto.CreateTagRequest{Name: "Duplicate"})
	if err == nil {
		t.Error("Expected error for duplicate tag name")
	}
	if err.Error() != "tag with this name already exists" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestTagService_CreateTag_DifferentUsers(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	// Same tag name is allowed for different users.
	_, err1 := svc.CreateTag(1, &dto.CreateTagRequest{Name: "Food"})
	_, err2 := svc.CreateTag(2, &dto.CreateTagRequest{Name: "Food"})
	if err1 != nil || err2 != nil {
		t.Errorf("Different users should be able to use the same tag name: %v, %v", err1, err2)
	}
}

func TestTagService_CreateTag_RepoError(t *testing.T) {
	repo := newMockTagRepo()
	repo.createError = errors.New("db error")
	svc := NewTagService(repo)

	_, err := svc.CreateTag(1, &dto.CreateTagRequest{Name: "Tag"})
	if err == nil {
		t.Error("Expected error from repository")
	}
}

func TestTagService_GetTags(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	svc.CreateTag(1, &dto.CreateTagRequest{Name: "A"})
	svc.CreateTag(1, &dto.CreateTagRequest{Name: "B"})
	svc.CreateTag(2, &dto.CreateTagRequest{Name: "C"}) // different user

	tags, err := svc.GetTags(1, "name")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags for user 1, got %d", len(tags))
	}
}

func TestTagService_GetTagByID(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	created, _ := svc.CreateTag(1, &dto.CreateTagRequest{Name: "Utilities"})
	tag, err := svc.GetTagByID(created.ID, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if tag.Name != "Utilities" {
		t.Errorf("Expected name 'Utilities', got '%s'", tag.Name)
	}
}

func TestTagService_GetTagByID_NotFound(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	_, err := svc.GetTagByID(999, 1)
	if err == nil {
		t.Error("Expected error for non-existent tag")
	}
}

func TestTagService_UpdateTag_Name(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	created, _ := svc.CreateTag(1, &dto.CreateTagRequest{Name: "Old"})
	newName := "New"
	updated, err := svc.UpdateTag(created.ID, 1, &dto.UpdateTagRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if updated.Name != "New" {
		t.Errorf("Expected name 'New', got '%s'", updated.Name)
	}
}

func TestTagService_UpdateTag_Color(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	created, _ := svc.CreateTag(1, &dto.CreateTagRequest{Name: "Tag"})
	newColor := "#AABBCC"
	updated, err := svc.UpdateTag(created.ID, 1, &dto.UpdateTagRequest{Color: &newColor})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if updated.Color != "#AABBCC" {
		t.Errorf("Expected color '#AABBCC', got '%s'", updated.Color)
	}
}

func TestTagService_UpdateTag_DuplicateName(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	svc.CreateTag(1, &dto.CreateTagRequest{Name: "Existing"})
	t2, _ := svc.CreateTag(1, &dto.CreateTagRequest{Name: "Other"})

	conflict := "Existing"
	_, err := svc.UpdateTag(t2.ID, 1, &dto.UpdateTagRequest{Name: &conflict})
	if err == nil {
		t.Error("Expected conflict error when renaming to existing name")
	}
}

func TestTagService_UpdateTag_NotFound(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	name := "X"
	_, err := svc.UpdateTag(999, 1, &dto.UpdateTagRequest{Name: &name})
	if err == nil {
		t.Error("Expected error for non-existent tag")
	}
}

func TestTagService_DeleteTag(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	created, _ := svc.CreateTag(1, &dto.CreateTagRequest{Name: "ToDelete"})
	if err := svc.DeleteTag(created.ID, 1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	_, err := svc.GetTagByID(created.ID, 1)
	if err == nil {
		t.Error("Expected error for deleted tag")
	}
}

func TestTagService_DeleteTag_RepoError(t *testing.T) {
	repo := newMockTagRepo()
	repo.deleteError = errors.New("delete failed")
	svc := NewTagService(repo)

	// Create a tag without hitting the delete error at creation.
	repo.deleteError = nil
	created, _ := svc.CreateTag(1, &dto.CreateTagRequest{Name: "Tag"})
	repo.deleteError = errors.New("delete failed")

	err := svc.DeleteTag(created.ID, 1)
	if err == nil {
		t.Error("Expected delete error")
	}
}

func TestTagService_SuggestTags_Empty(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	// No tags exist → should return empty slice, not error.
	suggestions, err := svc.SuggestTags(1, 5, "groceries at the market")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(suggestions) != 0 {
		t.Errorf("Expected 0 suggestions, got %d", len(suggestions))
	}
}

func TestTagService_SuggestTags_KeywordMatch(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	// Add tags where one matches the description keyword.
	repo.tags[1] = &models.Tag{ID: 1, UserID: 1, Name: "groceries", UsageCount: 5}
	repo.tags[2] = &models.Tag{ID: 2, UserID: 1, Name: "transport", UsageCount: 1}
	repo.nextID = 3

	suggestions, err := svc.SuggestTags(1, 5, "monthly groceries shopping")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// Should have at least the "groceries" tag suggested.
	if len(suggestions) == 0 {
		t.Error("Expected at least one suggestion for matching keyword")
	}
	if suggestions[0].Name != "groceries" {
		t.Errorf("Expected top suggestion 'groceries', got '%s'", suggestions[0].Name)
	}
}

func TestTagService_SuggestTags_AtMostFive(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	// Add more than 5 tags that all match.
	for i := uint(1); i <= 10; i++ {
		repo.tags[i] = &models.Tag{ID: i, UserID: 1, Name: "food", UsageCount: int(i)}
	}
	repo.nextID = 11

	suggestions, err := svc.SuggestTags(1, 5, "food expense at restaurant")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(suggestions) > 5 {
		t.Errorf("Expected at most 5 suggestions, got %d", len(suggestions))
	}
}

func TestTagService_GetSpendingByTag(t *testing.T) {
	repo := newMockTagRepo()
	svc := NewTagService(repo)

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	resp, err := svc.GetSpendingByTag(1, start, end)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.Period.StartDate != "2024-01-01" {
		t.Errorf("Expected start date '2024-01-01', got '%s'", resp.Period.StartDate)
	}
	if resp.Period.EndDate != "2024-01-31" {
		t.Errorf("Expected end date '2024-01-31', got '%s'", resp.Period.EndDate)
	}
}
