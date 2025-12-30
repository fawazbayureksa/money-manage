package controllers

import (
	"my-api/dto"
	"my-api/services"
	"my-api/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BankController struct {
	service services.BankService
}

func NewBankController(service services.BankService) *BankController {
	return &BankController{service: service}
}

// GetBanks godoc
// @Summary Get all banks with pagination and filtering
// @Description Get banks list with optional filters and pagination
// @Tags banks
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param search query string false "Search by bank name"
// @Param bank_name query string false "Filter by bank name"
// @Param color query string false "Filter by color"
// @Param sort_by query string false "Sort by field" default(id)
// @Param sort_dir query string false "Sort direction (asc/desc)" default(desc)
// @Success 200 {object} dto.PaginationResponse
// @Router /api/banks [get]
func (ctrl *BankController) GetBanks(c *gin.Context) {
	var filter dto.BankFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	result, err := ctrl.service.GetAllBanks(&filter)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Banks retrieved successfully", result)
}

// CreateBank godoc
// @Summary Create a new bank
// @Description Create a new bank with provided details
// @Tags banks
// @Accept json
// @Produce json
// @Param bank body dto.CreateBankRequest true "Bank data"
// @Success 201 {object} dto.BankResponse
// @Router /api/banks [post]
func (ctrl *BankController) CreateBank(c *gin.Context) {
	var req dto.CreateBankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input data")
		return
	}

	bank, err := ctrl.service.CreateBank(&req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Bank created successfully", bank)
}

// DeleteBank godoc
// @Summary Delete a bank
// @Description Delete bank by ID
// @Tags banks
// @Accept json
// @Produce json
// @Param id path int true "Bank ID"
// @Success 200
// @Router /api/banks/{id} [delete]
func (ctrl *BankController) DeleteBank(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid bank ID")
		return
	}

	if err := ctrl.service.DeleteBank(uint(id)); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Bank deleted successfully", nil)
}