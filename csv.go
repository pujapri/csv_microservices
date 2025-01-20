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
)

const batchSize = 5000 // Number of rows to process in a batch

func UploadCSV(c *gin.Context) {
	// Parse the uploaded file from the request
	file, err := c.FormFile("file")
	if err != nil {
		logger.Log.WithError(err).Error("Failed to get file from request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file"})
		return
	}

	// Save the uploaded file locally
	filePath := "./uploaded_" + file.Filename
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		logger.Log.WithError(err).Error("Failed to save uploaded file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Open the saved file
	csvFile, err := os.Open(filePath)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to open saved file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer csvFile.Close()

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

/*
// UploadCSV handles file upload and triggers CSV parsing
func UploadCSV(c *gin.Context) {
	fmt.Println("inside the uploaded function")

	// Retrieve the file from the request
	file, err := c.FormFile("file") // Ensure "file" matches the key in Postman's form-data
	if err != nil {
		logger.Log.WithError(err).Error("Failed to upload file")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file upload"})
		return
	}

	// Save the file locally with a unique name to avoid collisions
	filePath := "./uploaded_" + file.Filename
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		logger.Log.WithError(err).Error("Failed to save file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Parse and process CSV asynchronously
	go func() {
		if err := parseCSV(filePath); err != nil {
			logger.Log.WithError(err).Error("Error processing CSV")
		}
		// Optionally, delete the file after processing to save disk space
		if err := os.Remove(filePath); err != nil {
			logger.Log.WithError(err).Error("Failed to delete uploaded file")
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded and processing started"})
}

// parseCSV processes the CSV file and inserts data into the database
func parseCSV(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to open file")
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows := make(chan []string, 100)
	var wg sync.WaitGroup

	// Number of worker goroutines
	numWorkers := 1000
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for row := range rows {
				logger.Log.WithFields(map[string]interface{}{
					"worker": workerID,
					"row":    row,
				}).Info("Processing row")
				if err := db.InsertRecord(row); err != nil {
					logger.Log.WithError(err).Error("Failed to insert record into database")
				}
			}
		}(i)
	}

	// Read and send CSV rows to workers
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Log.WithError(err).Error("Error reading CSV")
			return err
		}
		rows <- row
	}

	close(rows) // Signal workers to stop
	wg.Wait()   // Wait for all workers to finish
	logger.Log.Info("CSV processing completed")
	return nil
}
*/
