package controllers

import (
	"my-api/dto"
	"my-api/services"
	"my-api/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AnalyticsController struct {
	service services.AnalyticsService
}

func NewAnalyticsController(service services.AnalyticsService) *AnalyticsController {
	return &AnalyticsController{service: service}
}

func (ctrl *AnalyticsController) GetSpendingByCategory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req dto.AnalyticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	result, err := ctrl.service.GetSpendingByCategory(userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Spending by category retrieved successfully", result)
}

func (ctrl *AnalyticsController) GetIncomeVsExpense(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req dto.AnalyticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	result, err := ctrl.service.GetIncomeVsExpense(userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Income vs expense retrieved successfully", result)
}

func (ctrl *AnalyticsController) GetTrendAnalysis(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req dto.AnalyticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	result, err := ctrl.service.GetTrendAnalysis(userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Trend analysis retrieved successfully", result)
}

func (ctrl *AnalyticsController) GetSpendingByBank(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req dto.AnalyticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	result, err := ctrl.service.GetSpendingByBank(userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Spending by bank retrieved successfully", result)
}

func (ctrl *AnalyticsController) GetSpendingByAsset(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req dto.AnalyticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	result, err := ctrl.service.GetSpendingByAsset(userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Spending by asset retrieved successfully", result)
}

func (ctrl *AnalyticsController) GetMonthlyComparison(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	months := 6
	if m := c.Query("months"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil {
			months = parsed
		}
	}

	var assetID *uint64
	if a := c.Query("asset_id"); a != "" {
		if parsed, err := strconv.ParseUint(a, 10, 64); err == nil {
			assetID = &parsed
		}
	}

	result, err := ctrl.service.GetMonthlyComparison(userID.(uint), months, assetID)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Monthly comparison retrieved successfully", result)
}

func (ctrl *AnalyticsController) GetDashboardSummary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse optional date parameters
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &parsed
		}
	}

	var assetID *uint64
	if a := c.Query("asset_id"); a != "" {
		if parsed, err := strconv.ParseUint(a, 10, 64); err == nil {
			assetID = &parsed
		}
	}

	result, err := ctrl.service.GetDashboardSummary(userID.(uint), startDate, endDate, assetID)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Dashboard summary retrieved successfully", result)
}

func (ctrl *AnalyticsController) GetYearlyReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	year := time.Now().Year()
	if y := c.Query("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}

	var assetID *uint64
	if a := c.Query("asset_id"); a != "" {
		if parsed, err := strconv.ParseUint(a, 10, 64); err == nil {
			assetID = &parsed
		}
	}

	result, err := ctrl.service.GetYearlyReport(userID.(uint), year, assetID)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Yearly report retrieved successfully", result)
}

func (ctrl *AnalyticsController) GetCategoryTrend(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	categoryID, err := strconv.ParseUint(c.Param("category_id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	var req dto.AnalyticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	result, err := ctrl.service.GetCategoryTrend(userID.(uint), uint(categoryID), &req)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Category trend retrieved successfully", result)
}
