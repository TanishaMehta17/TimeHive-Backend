package controller

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"
	"github.com/TanishaMehta17/TimeHive-Backend/config"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)


func generateRandomToken() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func sendEmail(to, subject, body string) error {
   err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	from := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	if from == "" || password == "" || host == "" || port == "" {
		return fmt.Errorf("SMTP environment variables not set")
	}

	addr := host + ":" + port

	msg := []byte(
		"From: " + from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
			body,
	)

	auth := smtp.PlainAuth("", from, password, host)

	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}


func SubmitGuardian(c *gin.Context) {
	var input struct {
		UserID        string `json:"user_id" binding:"required"`
		UserName      string `json:"user_name" binding:"required"`
		GuardianEmail string `json:"guardian_email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	db := config.DBConn


	token := generateRandomToken()
	expiresAt := time.Now().Add(24 * time.Hour)

	
	query := `
		INSERT INTO pending_guardians (user_id, guardian_email, token, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var pendingID string
	err := db.QueryRow(c, query, input.UserID, input.GuardianEmail, token, expiresAt).Scan(&pendingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pending guardian"})
		return
	}

	
	acceptLink := fmt.Sprintf("http://localhost:8080/verify-guardian?token=%s&action=accept", token)
	declineLink := fmt.Sprintf("http://localhost:8080/verify-guardian?token=%s&action=decline", token)

	
	subject := fmt.Sprintf("Guardian Invitation for %s", input.UserName)

	
	emailBody := fmt.Sprintf(`
		<html>
		<body>
			<p>Hello,</p>
			<p><b>%s</b> has invited you to be their guardian on <b>TimeHive</b>.</p>
			<p>Please choose an option below:</p>
			<a href="%s" style="background-color:green;color:white;padding:10px 20px;text-decoration:none;border-radius:5px;">Accept</a>
			&nbsp;&nbsp;
			<a href="%s" style="background-color:red;color:white;padding:10px 20px;text-decoration:none;border-radius:5px;">Decline</a>
			<p>This request will expire in 24 hours.</p>
		</body>
		</html>
	`, input.UserName, acceptLink, declineLink)

	
	if err := sendEmail(input.GuardianEmail, subject, emailBody); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation sent to guardian"})
}


func VerifyGuardian(c *gin.Context) {
	token := c.Query("token")
	action := c.Query("action") 

	if token == "" || action == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	db := config.DBConn

	
	var userID, guardianEmail string
	var expiresAt time.Time
	query := `
		SELECT user_id, guardian_email, expires_at
		FROM pending_guardians
		WHERE token = $1
	`
	err := db.QueryRow(c, query, token).Scan(&userID, &guardianEmail, &expiresAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
		return
	}

	if time.Now().After(expiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token expired"})
		return
	}

	if action == "accept" {
		
		updateQuery := `
			UPDATE "User"
			SET guardian_id = gen_random_uuid(), guardian_email = $1
			WHERE id = $2
		`
		_, err = db.Exec(c, updateQuery, guardianEmail, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user with guardian"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Guardian verified and added successfully"})
	} else if action == "decline" {
		c.JSON(http.StatusOK, gin.H{"message": "Guardian invitation declined"})
	}

	
	_, _ = db.Exec(c, `DELETE FROM pending_guardians WHERE token = $1`, token)
}
