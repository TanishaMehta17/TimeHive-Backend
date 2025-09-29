package routes

import (
	"github.com/TanishaMehta17/TimeHive-Backend/controller"
	"github.com/gin-gonic/gin"
)

func GuardianRoutes(rg *gin.RouterGroup){

	rg.POST("/makeguardian", controller.SubmitGuardian)
	rg.POST("/verifyguardian", controller.VerifyGuardian)
     
} 