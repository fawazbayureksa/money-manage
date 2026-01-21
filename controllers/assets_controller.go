package controllers

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "my-api/services"
    "my-api/utils"
)

type AssetController struct {
    service *services.AssetService
}

func NewAssetController(service *services.AssetService) *AssetController {
    return &AssetController{service: service}
}

func (ac *AssetController) ListAssets(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }
    
    assets, err := ac.service.ListAssets(userID.(uint))
    if err != nil {
        utils.JSONError(c, http.StatusBadRequest, err.Error())
        return
    }
    utils.JSONSuccess(c, "Assets retrieved successfully", assets)
}

func (ac *AssetController) GetAsset(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }
    
    idParam := c.Param("id")
    id, err := strconv.ParseUint(idParam, 10, 32)
    if err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Invalid asset ID")
        return
    }
    
    asset, err := ac.service.GetAsset(userID.(uint), uint(id))
    if err != nil {
        utils.JSONError(c, http.StatusBadRequest, err.Error())
        return
    }
    utils.JSONSuccess(c, "Asset retrieved successfully", asset)
}

func (ac *AssetController) CreateAsset(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }
    
    var dto services.CreateAssetDTO
    if err := c.ShouldBindJSON(&dto); err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Invalid JSON payload")
        return
    }
    
    asset, err := ac.service.CreateAsset(userID.(uint), dto)
    if err != nil {
        utils.JSONError(c, http.StatusBadRequest, err.Error())
        return
    }
    utils.JSONSuccess(c, "Asset created successfully", asset)
}

func (ac *AssetController) UpdateAsset(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }
    
    idParam := c.Param("id")
    id, err := strconv.ParseUint(idParam, 10, 32)
    if err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Invalid asset ID")
        return
    }
    
    var dto services.UpdateAssetDTO
    if err := c.ShouldBindJSON(&dto); err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Invalid JSON payload")
        return
    }
    
    asset, err := ac.service.UpdateAsset(userID.(uint), uint(id), dto)
    if err != nil {
        utils.JSONError(c, http.StatusBadRequest, err.Error())
        return
    }
    utils.JSONSuccess(c, "Asset updated successfully", asset)
}

func (ac *AssetController) DeleteAsset(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }
    
    idParam := c.Param("id")
    id, err := strconv.ParseUint(idParam, 10, 32)
    if err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Invalid asset ID")
        return
    }
    
    if err := ac.service.DeleteAsset(userID.(uint), uint(id)); err != nil {
        utils.JSONError(c, http.StatusBadRequest, err.Error())
        return
    }
    utils.JSONSuccess(c, "Asset deleted successfully", nil)
}

func (ac *AssetController) Summary(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }
    
    summary, err := ac.service.Summary(userID.(uint))
    if err != nil {
        utils.JSONError(c, http.StatusBadRequest, err.Error())
        return
    }
    utils.JSONSuccess(c, "Summary retrieved successfully", summary)
}
