package main

import (
	"analysis-api/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/analysis", controllers.AnalysisController)

	router.Run("0.0.0.0:8080")
}
