package controllers

import (
	"analysis-api/logging"
	"analysis-api/models"
	"analysis-api/services"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
)

var logger *slog.Logger

func AnalysisController(c *gin.Context) {
	logger = logging.Get()
	param_duration := c.Query("duration")
	dimension := c.Query("dimension")
	if len(param_duration) == 0 || len(dimension) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid parameters",
		})
		return
	}
	if !slices.Contains(models.DIMENSION_TYPE, dimension) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": dimension + " analysis is not supported",
		})
		return
	}

	duration, err := time.ParseDuration(param_duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Could not process the required duration " + param_duration,
			"error":   err.Error(),
		})
		return
	}

	watcher, err := services.StartAnalysis(dimension, c, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "something wrong happened while performing the analysis",
			"error":   err.Error(),
		})
		return
	}

	result := watcher.Finalize()
	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}
