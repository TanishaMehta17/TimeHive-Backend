package routes

import (
	"github.com/TanishaMehta17/TimeHive-Backend/controller"
	"github.com/gin-gonic/gin"
)


func AuthRoutes(rg *gin.RouterGroup) {
	rg.POST("/signup", controller.SignUp)
	rg.POST("/signin", controller.SignIn)
}
