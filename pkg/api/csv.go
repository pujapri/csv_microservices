package api

import (
	"csv-microservice/pkg/db"
	"csv-microservice/pkg/logger"
	"encoding/csv"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const batchSize = 5000 // Number of rows to process in a batch

// Logger interface that defines methods for logging
type Logger interface {
	Infof(format string, args ...interface{})
	WithError(err error) *logrus.Entry
	Debug(args ...interface{})
	Error(args ...interface{})
}

// DB interface for database operations, particularly InsertBatch
type DB interface {
	InsertBatch(batch [][]string) error
}

// CSVReader interface to abstract the CSV reading functionality
type CSVReader interface {
	Read() (record []string, err error)
}

func UploadCSV(c *gin.Context) {
	// Parse the uploaded file from the request
	file, err := c.FormFile("file")
	if err != nil {
		logger.Log.WithError(err).Error("Failed to get file from request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file"})
		return
	}

	// Log file upload attempt
	logger.Log.Infof("Received upload request for file: %s", file.Filename)

	// Save the uploaded file locally
	filePath := "./uploaded_" + file.Filename
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		logger.Log.WithError(err).Error("Failed to save uploaded file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Log the successful file upload
	logger.Log.Infof("Uploaded file saved successfully: %s", file.Filename)

	// Open the saved file
	csvFile, err := os.Open(filePath)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to open saved file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer csvFile.Close()

	logger.Log.Debug("Successfully opened the CSV file")

	reader := csv.NewReader(csvFile)
	var wg sync.WaitGroup

	// Create a buffered channel for batches
	batches := make(chan [][]string, 500)

	// Start worker goroutines
	numWorkers := 50
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for batch := range batches {
				start := time.Now()
				if err := db.InsertBatch(batch); err != nil {
					logger.Log.WithError(err).Errorf("Worker %d: Failed to insert batch", workerID)
				} else {
					logger.Log.Infof("Worker %d: Successfully inserted batch of %d records in %v", workerID, len(batch), time.Since(start))
				}
			}
		}(i)
	}

	// Read and group rows into batches
	var batch [][]string
	for {
		row, err := reader.Read()
		if err == io.EOF {
			if len(batch) > 0 {
				batches <- batch // Send the last batch
			}
			break
		}
		if err != nil {
			logger.Log.WithError(err).Error("Error reading CSV")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading CSV"})
			return
		}
		batch = append(batch, row)
		if len(batch) >= batchSize {
			batches <- batch
			batch = nil // Reset batch
		}
	}

	close(batches) // Close the channel to signal workers to stop
	wg.Wait()      // Wait for all workers to finish

	logger.Log.Info("CSV processing completed")
	c.JSON(http.StatusOK, gin.H{"message": "CSV processing completed"})
}
