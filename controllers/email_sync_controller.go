package controllers

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"my-api/services"
	"my-api/utils"
)

type EmailSyncController struct {
	emailSyncService services.EmailSyncService
}

func NewEmailSyncController(emailSyncService services.EmailSyncService) *EmailSyncController {
	return &EmailSyncController{emailSyncService: emailSyncService}
}

// GetAuthURL returns the Google OAuth2 authorization URL.
// GET /api/v2/email-sync/auth
func (ctrl *EmailSyncController) GetAuthURL(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	authURL, err := ctrl.emailSyncService.GetAuthURL(userID)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to generate auth URL: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Auth URL generated", gin.H{"auth_url": authURL}))
}

// HandleCallback handles the OAuth2 callback from Google.
// GET /api/v2/email-sync/callback  (public – no JWT required)
func (ctrl *EmailSyncController) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	oauthErr := c.Query("error")

	redirectBase := os.Getenv("EMAIL_SYNC_REDIRECT_URL")
	if redirectBase == "" {
		redirectBase = "/"
	}

	if oauthErr != "" {
		c.Redirect(http.StatusFound, redirectBase+"?email_sync=error&reason="+oauthErr)
		return
	}

	if code == "" || state == "" {
		c.Redirect(http.StatusFound, redirectBase+"?email_sync=error&reason=missing_params")
		return
	}

	if err := ctrl.emailSyncService.HandleCallback(code, state); err != nil {
		c.Redirect(http.StatusFound, redirectBase+"?email_sync=error&reason="+err.Error())
		return
	}

	c.Redirect(http.StatusFound, redirectBase+"?email_sync=success")
}

// SyncEmails triggers a manual email sync.
// POST /api/v2/email-sync/sync
func (ctrl *EmailSyncController) SyncEmails(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	result, err := ctrl.emailSyncService.SyncEmails(userID)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Sync completed", result))
}

// GetStatus returns connection status and recent sync logs.
// GET /api/v2/email-sync/status
func (ctrl *EmailSyncController) GetStatus(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	connected := ctrl.emailSyncService.IsConnected(userID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	logs, total, err := ctrl.emailSyncService.GetLogs(userID, page, limit)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to fetch sync logs")
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Status fetched", gin.H{
		"connected": connected,
		"logs":      logs,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total_items": total,
			"total_pages": totalPages,
		},
	}))
}

// Disconnect revokes the stored Gmail token.
// DELETE /api/v2/email-sync/disconnect
func (ctrl *EmailSyncController) Disconnect(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	if err := ctrl.emailSyncService.Disconnect(userID); err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to disconnect Gmail")
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Gmail disconnected successfully", nil))
}

// getUserID extracts and returns the authenticated user's ID from the Gin context.
func getUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse("User not authenticated"))
		return 0, false
	}
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Invalid user ID"))
		return 0, false
	}
	return uid, true
}
