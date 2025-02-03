package api

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func ensureLogFileExists(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create a default log file with sample entries
		sampleLogs := ``
		err = os.WriteFile(filePath, []byte(sampleLogs), 0644)
		if err != nil {
			return fmt.Errorf("failed to create default log file: %w", err)
		}
		log.WithField("filePath", filePath).Info("Default log file created")
	}
	return nil
}

func analyzeLogs(filePath string, level string, source string, rangeStart string, rangeEnd string) (map[string]int, error) {
	err := ensureLogFileExists(filePath)
	if err != nil {
		log.WithError(err).Error("Failed to ensure log file exists")
		return nil, err
	}

	logCounts := map[string]int{"INFO": 0, "ERROR": 0, "DEBUG": 0}

	file, err := os.Open(filePath)
	if err != nil {
		log.WithError(err).WithField("filePath", filePath).Error("Failed to open log file")
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 10*1024*1024) // 10 MB buffer
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		lineUpper := strings.ToUpper(line)

		// Apply level filter if specified
		if level != "" && !strings.Contains(lineUpper, strings.ToUpper(level)) {
			continue
		}

		// Apply source filter if specified
		if source != "" && !strings.Contains(line, source) {
			continue
		}

		// Apply date/time range filter if specified
		if rangeStart != "" || rangeEnd != "" {
			logTime, err := parseLogTimestamp(line)
			if err == nil {
				startTime, _ := time.Parse(time.RFC3339, rangeStart)
				endTime, _ := time.Parse(time.RFC3339, rangeEnd)

				if (rangeStart != "" && logTime.Before(startTime)) || (rangeEnd != "" && logTime.After(endTime)) {
					continue
				}
			}
		}

		// Count log levels
		if strings.Contains(lineUpper, "INFO") {
			logCounts["INFO"]++
		} else if strings.Contains(lineUpper, "ERROR") {
			logCounts["ERROR"]++
		} else if strings.Contains(lineUpper, "DEBUG") {
			logCounts["DEBUG"]++
		}
	}

	if err := scanner.Err(); err != nil {
		log.WithError(err).Error("Failed to read log file")
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	log.WithField("logCounts", logCounts).Info("Log analysis completed")
	return logCounts, nil
}

func parseLogTimestamp(line string) (time.Time, error) {
	fields := strings.Fields(line)
	if len(fields) > 0 {
		return time.Parse(time.RFC3339, fields[0])
	}
	return time.Time{}, fmt.Errorf("no timestamp found")
}

func getLogs(c *gin.Context) {
	level := c.DefaultQuery("level", "")
	source := c.DefaultQuery("source", "")
	rangeStart := c.DefaultQuery("start", "")
	rangeEnd := c.DefaultQuery("end", "")
	filePath := c.DefaultQuery("file", "File.log")

	logCounts, err := analyzeLogs(filePath, level, source, rangeStart, rangeEnd)
	if err != nil {
		log.WithError(err).Error("Failed to analyze logs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logCounts": logCounts,
		"filters":   gin.H{"level": level, "source": source, "start": rangeStart, "end": rangeEnd},
	})
}
