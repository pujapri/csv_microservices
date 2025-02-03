package db

import (
	"csv-microservice/pkg/logger"
	"fmt"
	"log"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Records represents the table structure
type Records struct {
	ID           uint `gorm:"primaryKey"`
	DeviceName   string
	DeviceType   string
	Brand        string
	Model        string
	OS           string
	OSVersion    string
	PurchaseDate string
	WarrantyEnd  string
	Status       string
	Price        uint
}

// ConnectDatabase initializes the database connection
func ConnectDatabase() error {
	var err error
	DSN := "host=db user=postgres password=password dbname=device_data port=5432 sslmode=disable"

	for i := 0; i < 10; i++ {
		DB, err = gorm.Open(postgres.Open(DSN), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Database connection failed. Retrying in 5 seconds... (Attempt %d/10)\n", i+1)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
		return err
	}

	err = DB.AutoMigrate(&Records{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
		return err
	}
	log.Println("Database connected and migrated successfully.")
	return nil
}

func InsertBatch(records [][]string) error {
	if len(records) < 2 { // Ensure there's data beyond the header
		logger.Log.Error("CSV file contains no data")
		return fmt.Errorf("CSV file contains no data")
	}

	// Skip the first row (header)
	dataRows := records[1:]

	var batch []Records
	for _, record := range dataRows {
		if len(record) != 11 {
			logger.Log.Errorf("Invalid record length: %+v", record)
			continue
		}

		// Ensure the price column is valid before converting
		priceStr := record[10]
		price, err := strconv.Atoi(priceStr)
		if err != nil {
			logger.Log.WithError(err).Errorf("Invalid price value: %s", priceStr)
			continue
		}

		newRecord := Records{
			DeviceName:   record[1],
			DeviceType:   record[2],
			Brand:        record[3],
			Model:        record[4],
			OS:           record[5],
			OSVersion:    record[6],
			PurchaseDate: record[7],
			WarrantyEnd:  record[8],
			Status:       record[9],
			Price:        uint(price),
		}
		batch = append(batch, newRecord)
	}

	// Use GORM to insert the batch
	err := DB.Create(&batch).Error
	if err != nil {
		logger.Log.WithError(err).Error("Failed to insert batch into database")
	} else {
		logger.Log.Infof("Successfully inserted batch of %d records", len(batch))
	}
	return err
}

func GetPaginatedRecords(page, limit int) ([]Records, error) {
	// Validate pagination parameters
	if page < 1 || limit < 1 {
		return nil, fmt.Errorf("invalid pagination parameters: page and limit must be greater than 0")
	}

	offset := (page - 1) * limit
	var records []Records
	result := DB.Limit(limit).Offset(offset).Find(&records)
	return records, result.Error
}
