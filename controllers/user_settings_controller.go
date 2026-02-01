package controllers

import (
	"my-api/dto"
	"my-api/services"
	"my-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserSettingsController struct {
	service services.UserSettingsService
}

func NewUserSettingsController(service services.UserSettingsService) *UserSettingsController {
	return &UserSettingsController{service: service}
}

// GetUserSettings godoc
// @Summary Get user's pay cycle settings
// @Description Get current authenticated user's pay cycle settings
// @Tags user-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserSettingsResponse
// @Router /api/user/settings [get]
func (ctrl *UserSettingsController) GetUserSettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	settings, err := ctrl.service.GetUserSettings(userID.(uint))
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "User settings retrieved successfully", settings)
}

// CreateUserSettings godoc
// @Summary Create user's pay cycle settings
// @Description Create pay cycle settings for the authenticated user
// @Tags user-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateUserSettingsRequest true "User settings data"
// @Success 201 {object} dto.UserSettingsResponse
// @Router /api/user/settings [post]
func (ctrl *UserSettingsController) CreateUserSettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req dto.CreateUserSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	settings, err := ctrl.service.CreateUserSettings(userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User settings created successfully",
		"data":    settings,
	})
}

// UpdateUserSettings godoc
// @Summary Update user's pay cycle settings
// @Description Update pay cycle settings for the authenticated user
// @Tags user-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateUserSettingsRequest true "User settings data"
// @Success 200 {object} dto.UserSettingsResponse
// @Router /api/user/settings [put]
func (ctrl *UserSettingsController) UpdateUserSettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req dto.UpdateUserSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	settings, err := ctrl.service.UpdateUserSettings(userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "User settings updated successfully", settings)
}

// DeleteUserSettings godoc
// @Summary Delete user's pay cycle settings
// @Description Delete pay cycle settings for the authenticated user (reset to defaults)
// @Tags user-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/user/settings [delete]
func (ctrl *UserSettingsController) DeleteUserSettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	err := ctrl.service.DeleteUserSettings(userID.(uint))
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "User settings deleted successfully", nil)
}
