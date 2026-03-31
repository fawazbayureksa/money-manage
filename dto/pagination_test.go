package dto

import (
	"testing"
)

func TestPaginationRequestSetDefaultsOnZero(t *testing.T) {
	p := &PaginationRequest{}
	p.SetDefaults()

	if p.Page != 1 {
		t.Errorf("Expected default page 1, got %d", p.Page)
	}
	if p.PageSize != 10 {
		t.Errorf("Expected default page_size 10, got %d", p.PageSize)
	}
	if p.SortDir != "desc" {
		t.Errorf("Expected default sort_dir 'desc', got '%s'", p.SortDir)
	}
}

func TestPaginationRequestSetDefaultsPreservesExisting(t *testing.T) {
	p := &PaginationRequest{Page: 3, PageSize: 25, SortDir: "asc"}
	p.SetDefaults()

	if p.Page != 3 {
		t.Errorf("Expected page 3, got %d", p.Page)
	}
	if p.PageSize != 25 {
		t.Errorf("Expected page_size 25, got %d", p.PageSize)
	}
	if p.SortDir != "asc" {
		t.Errorf("Expected sort_dir 'asc', got '%s'", p.SortDir)
	}
}

func TestPaginationRequestGetOffset(t *testing.T) {
	tests := []struct {
		page     int
		pageSize int
		expected int
	}{
		{1, 10, 0},
		{2, 10, 10},
		{3, 10, 20},
		{1, 20, 0},
		{2, 20, 20},
		{5, 5, 20},
	}

	for _, tt := range tests {
		p := &PaginationRequest{Page: tt.page, PageSize: tt.pageSize}
		offset := p.GetOffset()
		if offset != tt.expected {
			t.Errorf("Page=%d PageSize=%d: expected offset %d, got %d",
				tt.page, tt.pageSize, tt.expected, offset)
		}
	}
}

func TestNewPaginationResponseWithRemainder(t *testing.T) {
	resp := NewPaginationResponse([]string{"a", "b", "c"}, 1, 10, 25)

	if resp.Page != 1 {
		t.Errorf("Expected page 1, got %d", resp.Page)
	}
	if resp.PageSize != 10 {
		t.Errorf("Expected page_size 10, got %d", resp.PageSize)
	}
	if resp.TotalItems != 25 {
		t.Errorf("Expected total_items 25, got %d", resp.TotalItems)
	}
	if resp.TotalPages != 3 {
		t.Errorf("Expected total_pages 3, got %d", resp.TotalPages)
	}
}

func TestNewPaginationResponseExactDivision(t *testing.T) {
	resp := NewPaginationResponse(nil, 1, 10, 20)
	if resp.TotalPages != 2 {
		t.Errorf("Expected 2 pages, got %d", resp.TotalPages)
	}
}

func TestNewPaginationResponseSinglePage(t *testing.T) {
	resp := NewPaginationResponse(nil, 1, 10, 5)
	if resp.TotalPages != 1 {
		t.Errorf("Expected 1 page, got %d", resp.TotalPages)
	}
}

func TestNewPaginationResponseEmpty(t *testing.T) {
	resp := NewPaginationResponse(nil, 1, 10, 0)
	if resp.TotalPages != 0 {
		t.Errorf("Expected 0 pages, got %d", resp.TotalPages)
	}
	if resp.TotalItems != 0 {
		t.Errorf("Expected 0 total items, got %d", resp.TotalItems)
	}
}

func TestNewPaginationResponseDataPassthrough(t *testing.T) {
	data := []int{1, 2, 3}
	resp := NewPaginationResponse(data, 2, 5, 13)
	if resp.Data == nil {
		t.Error("Expected data to be non-nil")
	}
}
