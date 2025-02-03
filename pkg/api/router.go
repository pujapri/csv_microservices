package api

import (
	"github.com/gin-gonic/gin"
)

func StartServer() {
	r := gin.Default()

	r.POST("/upload", UploadCSV)
	r.GET("/records", GetRecords)
	r.GET("/logs", getLogs)

	r.Run(":8080")
}
