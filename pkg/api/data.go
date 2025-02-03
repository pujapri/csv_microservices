package api

import (
	"csv-microservice/pkg/db"
	"csv-microservice/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Fetch records from the database
func GetRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Fetch records
	records, err := db.GetPaginatedRecords(page, limit)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to fetch records")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch records"})
		return
	}

	// Log number of records returned
	logger.Log.Infof("Fetched %d records", len(records))

	// Respond with the records
	c.JSON(http.StatusOK, gin.H{"records": records})
}
