package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Event struct {
	ID             string  `json:"id"`
	ProjectID      string  `json:"project_id"`
	Level          string  `json:"level"`
	Message        string  `json:"message"`
	StackTrace     *string `json:"stack_trace"`
	EventTimestamp string  `json:"event_timestamp"`
}

func parseDuration(param string)(time.Duration, error){
	return time.ParseDuration(param)
}

func GetEvents(db *gorm.DB) gin.HandlerFunc{
	return func(c * gin.Context){
		var events Event
		projectID := c.Param("id")

		result := db.Where("project_id = ?", projectID).Order("event_timestamp DESC").Limit(50).Find(&events)

		if result.Error != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return 
		}

		c.JSON(http.StatusOK, events)
	}
}

func GetErrorCount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var count int64
		projectID := c.Param("id")
		
		result := db.Model(&Event{}).Where("project_id = ? AND level = ?", projectID, "error").Count(&count)

		if result.Error != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return 
		}
		c.JSON(http.StatusOK, gin.H{"error_count": count})
	}
}

func GetErrorRate(db *gorm.DB) gin.HandlerFunc{
		return func(c *gin.Context) {
		projectID := c.Param("id")
		last := c.DefaultQuery("last", "5m")

		duration, err := time.ParseDuration(last)
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid duration format"})
			return
		}

		since := time.Now().Add(-duration)
		var total int64
		var errors int64
		var rate float64

		
		db.Model(&Event{}).Where("project_id = ? AND event_timestamp >= ?", projectID, since).Count(&total)

		db.Model(&Event{}).Where("project_id = ? AND level = ? AND event_timestamp >= ?", projectID, "error", since).Count(&errors)

		if rate > 0{
			rate = (float64(errors) / float64(total)) * 100
		}

		c.JSON(http.StatusOK, gin.H{
			"window":      last,
			"total":       total,
			"errors":      errors,
			"error_rate":  rate,
		})
}
}