package controllers

import (
    "errors"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "my-api/models"
    "my-api/services"
)

type AssetController struct {
    service *services.AssetService
}

func NewAssetController(service *services.AssetService) *AssetController {
    return &AssetController{service: service}
}

// Helper to fetch current user id from context
func getCurrentUserID(c *gin.Context) (uint64, error) {
    // Primary: userID in context
    if v, ok := c.Get("userID"); ok {
        switch val := v.(type) {
        case uint64:
            return val, nil
        case int:
            return uint64(val), nil
        case int64:
            return uint64(val), nil
        }
    }
    // Fallback: common alternative keys
    if v, ok := c.Get("user_id"); ok {
        switch val := v.(type) {
        case uint64:
            return val, nil
        case int:
            return uint64(val), nil
        case int64:
            return uint64(val), nil
        }
    }
    // Optional: if a User object is attached to context
    if v, ok := c.Get("currentUser"); ok {
        switch u := v.(type) {
        case *models.User:
            return uint64(u.ID), nil
        case models.User:
            return uint64(u.ID), nil
        }
    }
    if v, ok := c.Get("user"); ok {
        if u, ok := v.(models.User); ok {
            return uint64(u.ID), nil
        }
    }
    return 0, errors.New("unauthorized")
}

func (ac *AssetController) ListAssets(c *gin.Context) {
    userID, err := getCurrentUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }
    assets, err := ac.service.ListAssets(userID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, assets)
}

func (ac *AssetController) GetAsset(c *gin.Context) {
    userID, err := getCurrentUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }
    idParam := c.Param("id")
    id, err := strconv.ParseUint(idParam, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    asset, err := ac.service.GetAsset(userID, id)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, asset)
}

func (ac *AssetController) CreateAsset(c *gin.Context) {
    userID, err := getCurrentUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }
    var dto services.CreateAssetDTO
    if err := c.ShouldBindJSON(&dto); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
        return
    }
    asset, err := ac.service.CreateAsset(userID, dto)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, asset)
}

func (ac *AssetController) UpdateAsset(c *gin.Context) {
    userID, err := getCurrentUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }
    idParam := c.Param("id")
    id, err := strconv.ParseUint(idParam, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    var dto services.UpdateAssetDTO
    if err := c.ShouldBindJSON(&dto); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
        return
    }
    asset, err := ac.service.UpdateAsset(userID, id, dto)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, asset)
}

func (ac *AssetController) DeleteAsset(c *gin.Context) {
    userID, err := getCurrentUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }
    idParam := c.Param("id")
    id, err := strconv.ParseUint(idParam, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    if err := ac.service.DeleteAsset(userID, id); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.Status(http.StatusNoContent)
}

func (ac *AssetController) Summary(c *gin.Context) {
    userID, err := getCurrentUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }
    summary, err := ac.service.Summary(userID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, summary)
}
