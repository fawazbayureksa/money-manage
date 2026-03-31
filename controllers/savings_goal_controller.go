package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"my-api/dto"
	"my-api/services"
	"my-api/utils"
)

type SavingsGoalController struct {
	service services.SavingsGoalService
}

func NewSavingsGoalController(service services.SavingsGoalService) *SavingsGoalController {
	return &SavingsGoalController{service: service}
}

func (ctrl *SavingsGoalController) CreateGoal(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req dto.CreateSavingsGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input data: "+err.Error())
		return
	}

	goal, err := ctrl.service.CreateGoal(userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Savings goal created successfully", goal)
}

func (ctrl *SavingsGoalController) GetGoal(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid goal ID")
		return
	}

	goal, err := ctrl.service.GetGoalByID(uint(id), userID.(uint))
	if err != nil {
		utils.JSONError(c, http.StatusNotFound, err.Error())
		return
	}

	utils.JSONSuccess(c, "Savings goal retrieved successfully", goal)
}

func (ctrl *SavingsGoalController) GetGoals(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var filter dto.SavingsGoalFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	result, err := ctrl.service.GetAllGoals(userID.(uint), &filter)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Savings goals retrieved successfully", result)
}

func (ctrl *SavingsGoalController) UpdateGoal(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid goal ID")
		return
	}

	var req dto.UpdateSavingsGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input data: "+err.Error())
		return
	}

	goal, err := ctrl.service.UpdateGoal(uint(id), userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Savings goal updated successfully", goal)
}

func (ctrl *SavingsGoalController) DeleteGoal(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid goal ID")
		return
	}

	if err := ctrl.service.DeleteGoal(uint(id), userID.(uint)); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Savings goal deleted successfully", nil)
}

func (ctrl *SavingsGoalController) AddContribution(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	goalID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid goal ID")
		return
	}

	var req dto.AddContributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input data: "+err.Error())
		return
	}

	contribution, err := ctrl.service.AddContribution(uint(goalID), userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Contribution added successfully", contribution)
}

func (ctrl *SavingsGoalController) GetContributions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	goalID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid goal ID")
		return
	}

	var filter dto.ContributionFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	result, err := ctrl.service.GetContributions(uint(goalID), userID.(uint), &filter)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Contributions retrieved successfully", result)
}

func (ctrl *SavingsGoalController) DeleteContribution(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	goalID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid goal ID")
		return
	}

	contributionID, err := strconv.ParseUint(c.Param("contribution_id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid contribution ID")
		return
	}

	if err := ctrl.service.DeleteContribution(uint(contributionID), uint(goalID), userID.(uint)); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Contribution deleted successfully", nil)
}
