package routes

import (
	"crypto-tracker/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupAlertRoutes configura tutte le routes per gli alert
func SetupAlertRoutes(router *gin.Engine, db *gorm.DB) {
	alertRoutes := router.Group("/alerts")
	{
		alertRoutes.GET("/", controllers.GetAlerts(db))
		alertRoutes.GET("/active", controllers.GetActiveAlerts(db))
		alertRoutes.GET("/:id", controllers.GetAlert(db))
		alertRoutes.POST("/", controllers.CreateAlert(db))
		alertRoutes.PUT("/:id", controllers.UpdateAlert(db))
		alertRoutes.DELETE("/:id", controllers.DeleteAlert(db))
	}
}
