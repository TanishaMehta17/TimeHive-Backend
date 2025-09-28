package controller

import(
	"net/http"	
	"github.com/TanishaMehta17/TimeHive-Backend/config"
	"github.com/TanishaMehta17/TimeHive-Backend/model"
	"github.com/gin-gonic/gin" 
)

func MakeGuardian(c *gin.Context) {
    var input struct {
        UserID        string `json:"user_id" binding:"required"`
        GuardianEmail string `json:"guardian_name" binding:"required"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format, user_id and guardian email required"})
        return
    }

    db := config.DBConn
    var user model.User

    query := `
        UPDATE "User" 
        SET guardian_id = gen_random_uuid(), guardian_email = $1
        WHERE user_id = $2
        RETURNING guardian_id, guardian_email
    `
    err := db.QueryRow(c, query, input.GuardianEmail, input.UserID).
        Scan(&user.GuardianID, &user.GuardianEmail)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update guardian"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Guardian added successfully", "guardian": user})
}
