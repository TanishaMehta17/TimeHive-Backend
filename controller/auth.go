package controller

import (
	"github.com/TanishaMehta17/TimeHive-Backend/config"
	"github.com/TanishaMehta17/TimeHive-Backend/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWT Secret (set via env for production)
var jwtSecret = []byte("mysecretkey") // replace with env var in production

func AuthRoutes(r *gin.RouterGroup) {
	r.POST("/signup", SignUp)
	r.POST("/signin", SignIn)
}

// SignUp API
func SignUp(c *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user in DB
	db := config.DBConn
	query := `INSERT INTO "User" (user_id, name, email, password, created_at)
			  VALUES (gen_random_uuid(), $1, $2, $3, NOW()) RETURNING user_id, name, email, created_at`
	var user model.User
	err = db.QueryRow(c, query, input.Name, input.Email, string(hashedPassword)).
		Scan(&user.UserID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user": user})
}

// SignIn API
func SignIn(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch user from DB
	db := config.DBConn
	var user model.User
	var hashedPassword string
	query := `SELECT user_id, name, email, password, created_at FROM "User" WHERE email=$1`
	err := db.QueryRow(c, query, input.Email).
		Scan(&user.UserID, &user.Name, &user.Email, &hashedPassword, &user.CreatedAt)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.UserID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // token valid 24h
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": tokenString})
}
