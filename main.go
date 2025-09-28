package main

import (
	"fmt"

	"github.com/TanishaMehta17/TimeHive-Backend/config"
	"github.com/TanishaMehta17/TimeHive-Backend/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	db := config.ConnectDB()

	if db != nil {
		fmt.Println(" Connected to PostgreSQL successfully!")
	} else {
		fmt.Println(" Failed to connect to PostgreSQL.")
	}
	router := gin.Default()
	authGroup := router.Group("/api/auth")
	guardianGroup := router.Group("/api/guardian")
	

	routes.AuthRoutes(authGroup)
	routes.GuardianRoutes(guardianGroup)
	
	router.Run(":8080")

}
