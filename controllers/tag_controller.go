package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"my-api/dto"
	"my-api/services"
	"my-api/utils"
)

type TagController struct {
	service services.TagService
}

// NewTagController creates a new tag controller
func NewTagController(service services.TagService) *TagController {
	return &TagController{service: service}
}

// CreateTag godoc
// @Summary Create a new tag
// @Description Create a new tag for the authenticated user
// @Tags tags
// @Accept json
// @Produce json
// @Param tag body dto.CreateTagRequest true "Tag data"
// @Success 201 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v2/tags [post]
func (tc *TagController) CreateTag(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse("User not authenticated"))
		return
	}

	var req dto.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
		return
	}

	tag, err := tc.service.CreateTag(userID.(uint), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse("Tag created successfully", tag))
}

// GetTags godoc
// @Summary Get all tags
// @Description Get all tags for the authenticated user
// @Tags tags
// @Produce json
// @Param sort query string false "Sort by (usage or name)" default(name)
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v2/tags [get]
func (tc *TagController) GetTags(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse("User not authenticated"))
		return
	}

	sortBy := c.DefaultQuery("sort", "name")

	tags, err := tc.service.GetTags(userID.(uint), sortBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Tags fetched successfully", tags))
}

// GetTagByID godoc
// @Summary Get tag by ID
// @Description Get a specific tag by ID
// @Tags tags
// @Produce json
// @Param id path int true "Tag ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v2/tags/{id} [get]
func (tc *TagController) GetTagByID(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse("User not authenticated"))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid tag ID"))
		return
	}

	tag, err := tc.service.GetTagByID(uint(id), userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Tag fetched successfully", tag))
}

// UpdateTag godoc
// @Summary Update a tag
// @Description Update an existing tag
// @Tags tags
// @Accept json
// @Produce json
// @Param id path int true "Tag ID"
// @Param tag body dto.UpdateTagRequest true "Updated tag data"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v2/tags/{id} [put]
func (tc *TagController) UpdateTag(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse("User not authenticated"))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid tag ID"))
		return
	}

	var req dto.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
		return
	}

	tag, err := tc.service.UpdateTag(uint(id), userID.(uint), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Tag updated successfully", tag))
}

// DeleteTag godoc
// @Summary Delete a tag
// @Description Delete a tag (soft delete)
// @Tags tags
// @Produce json
// @Param id path int true "Tag ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v2/tags/{id} [delete]
func (tc *TagController) DeleteTag(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse("User not authenticated"))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid tag ID"))
		return
	}

	err = tc.service.DeleteTag(uint(id), userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Tag deleted successfully", nil))
}

// SuggestTags godoc
// @Summary Suggest tags
// @Description Get tag suggestions based on category and description
// @Tags tags
// @Produce json
// @Param category_id query int true "Category ID"
// @Param description query string false "Transaction description"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v2/tags/suggest [get]
func (tc *TagController) SuggestTags(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse("User not authenticated"))
		return
	}

	categoryIDStr := c.Query("category_id")
	if categoryIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("category_id is required"))
		return
	}

	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid category_id"))
		return
	}

	description := c.DefaultQuery("description", "")

	suggestions, err := tc.service.SuggestTags(userID.(uint), uint(categoryID), description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Tag suggestions fetched successfully", suggestions))
}

// GetSpendingByTag godoc
// @Summary Get spending by tag
// @Description Get spending analytics grouped by tags
// @Tags tags
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v2/analytics/spending-by-tag [get]
func (tc *TagController) GetSpendingByTag(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse("User not authenticated"))
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("start_date and end_date are required"))
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid start_date format (use YYYY-MM-DD)"))
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid end_date format (use YYYY-MM-DD)"))
		return
	}

	response, err := tc.service.GetSpendingByTag(userID.(uint), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Spending by tag fetched successfully", response))
}
