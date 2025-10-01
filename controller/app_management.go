package controller

import (
	"context"
	"fmt"
	"github.com/TanishaMehta17/TimeHive-Backend/config"

	"github.com/TanishaMehta17/TimeHive-Backend/utility"
	"github.com/gin-gonic/gin"
)

// fetch guardian email for user
func getGuardianEmail(userID string) string {
	var email string
	db := config.DBConn
	err := db.QueryRow(context.Background(), `SELECT guardian_email FROM "User" WHERE user_id=$1`, userID).Scan(&email)
	if err != nil {
		return ""
	}
	return email
}

// fetch user name
func getUserName(userID string) string {
	var name string
	db := config.DBConn
	err := db.QueryRow(context.Background(), `SELECT name FROM "User" WHERE user_id=$1`, userID).Scan(&name)
	if err != nil {
		return "User"
	}
	return name
}

// ---------------- Set App Limit ----------------
func SetAppLimit(c *gin.Context) {
	var input struct {
		UserID       string `json:"user_id" binding:"required"`
		AppID        string `json:"app_id" binding:"required"`
		LimitMinutes int    `json:"limit_minutes" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	db := config.DBConn
	query := `
		INSERT INTO "AppLimit" (user_id, app_id, daily_limit_minutes, created_at)
		VALUES ($1,$2,$3,NOW())
		ON CONFLICT (user_id, app_id)
		DO UPDATE SET daily_limit_minutes=$3
		RETURNING limit_id
	`
	var limitID int
	err := db.QueryRow(c, query, input.UserID, input.AppID, input.LimitMinutes).Scan(&limitID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to set app limit"})
		return
	}

	guardianEmail := getGuardianEmail(input.UserID)
	userName := getUserName(input.UserID)
	if guardianEmail != "" {
		subject := fmt.Sprintf("App Limit Set for AppID: %s", input.AppID)
		body := fmt.Sprintf(`<p>%s set a daily limit of <strong>%d minutes</strong> for app <strong>%s</strong>.</p>
			<p>You will be notified if the user exceeds this limit.</p>`, userName, input.LimitMinutes, input.AppID)
		utility.SendEmail(guardianEmail, subject, body)
	}

	c.JSON(200, gin.H{"message": "App limit set and guardian notified"})
}

// ---------------- Track App Usage ----------------
func TrackAppUsage(c *gin.Context) {
	var input struct {
		UserID     string `json:"user_id" binding:"required"`
		AppID      string `json:"app_id" binding:"required"`
		DeviceID   string `json:"device_id" binding:"required"`
		MinutesUsed int   `json:"minutes_used" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	db := config.DBConn
	//today := time.Now().Format("2006-01-02")

	// Insert session per device
	_, _ = db.Exec(c, `
		INSERT INTO "AppUsageSession" (user_id, app_id, device_id, session_start, duration_minutes)
		VALUES ($1,$2,$3,NOW(),$4)
	`, input.UserID, input.AppID, input.DeviceID, input.MinutesUsed)

	// Update daily usage
	_, _ = db.Exec(c, `
		INSERT INTO "DailyAppUsage" (user_id, app_id, usage_date, total_minutes, blocked, updated_at)
		VALUES ($1,$2,CURRENT_DATE,$3,false,NOW())
		ON CONFLICT (user_id, app_id, usage_date)
		DO UPDATE SET total_minutes=DailyAppUsage.total_minutes+$3, updated_at=NOW()
	`, input.UserID, input.AppID, input.MinutesUsed)

	// Get total usage and limit
	var totalUsed, dailyLimit int
	db.QueryRow(c, `SELECT total_minutes FROM "DailyAppUsage" WHERE user_id=$1 AND app_id=$2 AND usage_date=CURRENT_DATE`, input.UserID, input.AppID).Scan(&totalUsed)
	db.QueryRow(c, `SELECT daily_limit_minutes FROM "AppLimit" WHERE user_id=$1 AND app_id=$2`, input.UserID, input.AppID).Scan(&dailyLimit)

	// Check if limit exceeded and guardian override is needed
	if totalUsed > dailyLimit {
		var blocked bool
		db.QueryRow(c, `SELECT blocked FROM "DailyAppUsage" WHERE user_id=$1 AND app_id=$2 AND usage_date=CURRENT_DATE`, input.UserID, input.AppID).Scan(&blocked)
		if !blocked {
			guardianEmail := getGuardianEmail(input.UserID)
			userName := getUserName(input.UserID)
			token := utility.GenerateRandomToken()
			// Save guardian override
			_, _ = db.Exec(c, `INSERT INTO guardian_overrides (user_id, app_id, token, extra_minutes, valid_for)
				VALUES ($1,$2,$3,$4,CURRENT_DATE)`, input.UserID, input.AppID, token, 30)

			acceptLink := fmt.Sprintf("http://localhost:8080/guardian-action?token=%s&action=accept", token)
			declineLink := fmt.Sprintf("http://localhost:8080/guardian-action?token=%s&action=decline", token)
			subject := fmt.Sprintf("Limit Exceeded for %s", input.AppID)
			body := fmt.Sprintf(`
				<p>%s exceeded daily limit of <strong>%d min</strong> for app <strong>%s</strong>.</p>
				<p>Approve additional 30 min for today?</p>
				<a href="%s" style="background:green;color:white;padding:8px 12px;text-decoration:none;">Accept</a>
				<a href="%s" style="background:red;color:white;padding:8px 12px;text-decoration:none;">Decline</a>
			`, userName, dailyLimit, input.AppID, acceptLink, declineLink)
			utility.SendEmail(guardianEmail, subject, body)
		}
	}

	c.JSON(200, gin.H{"message": "Usage updated"})
}

// ---------------- Guardian Action ----------------
func GuardianAction(c *gin.Context) {
	token := c.Query("token")
	action := c.Query("action")
	db := config.DBConn
	//today := time.Now().Format("2006-01-02")

	var userID, appID string
	var extraMinutes int
	err := db.QueryRow(c, `SELECT user_id, app_id, extra_minutes FROM guardian_overrides WHERE token=$1 AND valid_for=CURRENT_DATE`, token).Scan(&userID, &appID, &extraMinutes)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid or expired token"})
		return
	}

	if action == "accept" {
		// Add extra minutes to daily usage limit for today
		_, _ = db.Exec(c, `
			UPDATE "DailyAppUsage"
			SET total_minutes = total_minutes - $3, updated_at=NOW()
			WHERE user_id=$1 AND app_id=$2 AND usage_date=CURRENT_DATE
		`, userID, appID, extraMinutes)
		c.JSON(200, gin.H{"message": "Guardian approved extra time"})
	} else if action == "decline" {
		// Block app for the rest of today
		_, _ = db.Exec(c, `
			UPDATE "DailyAppUsage"
			SET blocked=true, updated_at=NOW()
			WHERE user_id=$1 AND app_id=$2 AND usage_date=CURRENT_DATE
		`, userID, appID)
		c.JSON(200, gin.H{"message": "Guardian declined extension"})
	}

	// Consume token
	_, _ = db.Exec(c, `DELETE FROM guardian_overrides WHERE token=$1`, token)
}
