package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/k1ngalph0x/beacon/services/issue-service/models"
	publisher "github.com/k1ngalph0x/beacon/services/issue-service/utils"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)




type IssueUpdateEvent struct {
	IssueID   string    `json:"issue_id"`
	ProjectID string    `json:"project_id"`
	Count     int       `json:"count"`
	Level     string    `json:"level"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type IssueResolvedEvent struct {
	IssueID   string    `json:"issue_id"`
	ProjectID string    `json:"project_id"`
	ResolvedAt time.Time `json:"resolved_at"`
}


var issueResolvedWriter = &kafka.Writer{
	Addr:     kafka.TCP("localhost:9092"),
	Topic:    "issue-resolved",
	Balancer: &kafka.LeastBytes{},
}


var issueUpdateWriter = &kafka.Writer{
	Addr:     kafka.TCP("localhost:9092"),
	Topic:    "issue-updates",
	Balancer: &kafka.LeastBytes{},
}


func generateFingerprint(message string, stack *string) string{
	raw := message
	if stack != nil{
		raw += "|" + *stack
	}

	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}


func verifyProjectOwnership(db *gorm.DB, userID, projectID string) bool{
	var count int64

	result := db.Table("projects").Where("id = ? AND user_id = ?", projectID, userID).Count(&count)

	if result.Error  != nil{
		return false
	}

	return count > 0
}

func GetProjectIssue(db *gorm.DB) gin.HandlerFunc{
	return func(c *gin.Context){
		var issues []models.Issue
		
		projectID := c.Param("project_id")
		userID := c.GetString("user_id")

		if userID == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		if !verifyProjectOwnership(db, userID, projectID){
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}

		err := db.Where("project_id = ?", projectID).Order("last_seen desc").Find(&issues)
		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(200, gin.H{"issues": issues}) 
	}
}


func ResolveIssue(db *gorm.DB) gin.HandlerFunc{
	return func(c *gin.Context){
		var issue models.Issue
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		
		id := c.Param("id")
		err := db.First(&issue, "id = ?", id).Error; 
		if err != nil {
			c.JSON(404, gin.H{"error": "Issue not found"})
			return
		}

		if !verifyProjectOwnership(db, userID, issue.ProjectID) {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}

		issue.Status = "resolved"
		result := db.Save(&issue)
		if result.Error != nil{
			c.JSON(500, gin.H{"error": "Failed to update issue"})
			return
		}

		resolvedEvent := IssueResolvedEvent{
			IssueID:   issue.ID,
			ProjectID: issue.ProjectID,	
			ResolvedAt: time.Now(),
		}

		publisher.PublishEvent(issueResolvedWriter, issue.ProjectID, resolvedEvent)

		c.JSON(200, gin.H{
			"message": "Issue resolved",
			"issue": issue,
		})
	}
}


func GetIssue(db *gorm.DB) gin.HandlerFunc{
	return func(c *gin.Context){
		id := c.Param("id")

		var issue models.Issue

		result := db.First(&issue, "id = ?", id); 
		if result.Error != nil {
			c.JSON(404, gin.H{"error": "Issue not found"})
			return
		}

		c.JSON(200, gin.H{
			"message": "Issue resolved",
			"issue": issue,
		})
	}
}

func ProcessEvent(conn *gorm.DB, e models.Event){
	var issue models.Issue
	fp := generateFingerprint(e.Message, e.StackTrace)

	err := conn.Where("project_id = ? AND fingerprint = ?", e.ProjectID, fp).First(&issue).Error
	if err == nil{
		err := conn.Model(&issue).Updates(map[string]interface{}{
			"count":     gorm.Expr("count + ?", 1),
			"last_seen": time.Now(),
		}).Error

		if err != nil {
			log.Printf("Failed to update issue %s: %v", issue.ID, err)
		}

		conn.First(&issue, "id = ?", issue.ID)

		updateEvent := IssueUpdateEvent{
			IssueID:   issue.ID,
			ProjectID: issue.ProjectID,
			Count:     issue.Count,
			Level:     issue.Level,
			Status:    issue.Status,
			UpdatedAt: time.Now(),
		}

		publisher.PublishEvent(issueUpdateWriter, issue.ProjectID, updateEvent)

		return
	}

	newIssue := models.Issue{
		ID:          uuid.New().String(),
		ProjectID:   e.ProjectID,
		Fingerprint: fp,
		Title:       e.Message,
		Level:       e.Level,
		Count:       1,
		FirstSeen:   time.Now(),
		LastSeen:    time.Now(),
		Status:      "open",
	}

	err = conn.Create(&newIssue).Error
	if err != nil{
		log.Println("Failed to create issue:", err)
		return
	}

	updateEvent := IssueUpdateEvent{
		IssueID:   newIssue.ID,
		ProjectID: newIssue.ProjectID,
		Count:     newIssue.Count,
		Level:     newIssue.Level,
		Status:    newIssue.Status,
		UpdatedAt: time.Now(),
	}

	publisher.PublishEvent(issueUpdateWriter, newIssue.ProjectID, updateEvent)
}
