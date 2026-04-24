package job

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
	"time"
)

type JobStatus string

const (
	Pending   JobStatus = "Pending"
	Running   JobStatus = "Running"
	Completed JobStatus = "Completed"
	Failed    JobStatus = "Failed"
)

type Job struct {
	ID         uint      `gorm:"primaryKey"`
	TenantID   string    `gorm:"index"`
	JobType    string    `gorm:"index"`
	Payload    string    `gorm:"type:text"`
	Status     JobStatus `gorm:"index"`
	LockedBy   string    `gorm:"index"`
	LockedAt   *time.Time
	ErrorLog   string    `gorm:"type:text"`
	RetryCount int       `gorm:"default:0"`
	MaxRetries int       `gorm:"default:3"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

var nodeID string

func init() {
	hostname, _ := os.Hostname()
	nodeID = fmt.Sprintf("%s-%d", hostname, os.Getpid())
}

// HandlerFunc adalah fungsi yang akan memproses job
type HandlerFunc func(ctx context.Context, tenantID string, payload map[string]interface{}) error

var handlers = make(map[string]HandlerFunc)

func RegisterHandler(jobType string, handler HandlerFunc) {
	handlers[jobType] = handler
}

// Enqueue memasukkan job baru ke dalam antrean
func Enqueue(tenantID string, jobType string, payload map[string]interface{}) error {
	payloadBytes, _ := json.Marshal(payload)
	job := Job{
		TenantID:   tenantID,
		JobType:    jobType,
		Payload:    string(payloadBytes),
		Status:     Pending,
		MaxRetries: 3,
	}
	return database.DB.Table("tabJob").Create(&job).Error
}

// StartWorker menjalankan pemroses job di background
func StartWorker(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second) // Check more frequently
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			processJobs(ctx)
		}
	}
}

func processJobs(ctx context.Context) {
	// 1. Atomic Pick-up menggunakan PostgreSQL 'SKIP LOCKED'
	// Ini memastikan antar instance tidak berebut job yang sama
	var j Job
	query := `
		UPDATE "tabJob" 
		SET status = ?, locked_by = ?, locked_at = NOW(), updated_at = NOW()
		WHERE id = (
			SELECT id FROM "tabJob"
			WHERE (status = ? OR (status = ? AND retry_count < max_retries))
			AND (locked_at IS NULL OR locked_at < NOW() - INTERVAL '10 minutes')
			ORDER BY id ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING *
	`

	res := database.DB.Raw(query, Running, nodeID, Pending, Failed).Scan(&j)
	if res.Error != nil || res.RowsAffected == 0 {
		return // Tidak ada job yang siap
	}

	executeJob(ctx, j)
}
func executeJob(ctx context.Context, j Job) {
	executionID := fmt.Sprintf("EXE-%d-%s", j.ID, nodeID)
	handler, ok := handlers[j.JobType]
	if !ok {
		updateJobStatus(j.ID, Failed, "No handler registered")
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		logEntry := map[string]interface{}{
			"job_id": j.ID, "execution_id": executionID, "tenant_id": j.TenantID, "started_at": time.Now(),
		}
		if err := tx.Table("tabJobLog").Create(logEntry).Error; err != nil {
			return fmt.Errorf("idempotency violation")
		}

		var payload map[string]interface{}
		json.Unmarshal([]byte(j.Payload), &payload)
		return handler(ctx, j.TenantID, payload)
	})

	if err != nil {
		newRetryCount := j.RetryCount + 1
		status := Failed

		// V2: Exponential Backoff Logic
		// Next retry: 2^retry_count * 30 seconds
		backoffSeconds := time.Duration(math.Pow(2, float64(newRetryCount))) * 30
		nextRetryAt := time.Now().Add(backoffSeconds * time.Second)

		if newRetryCount >= j.MaxRetries {
			// Move to Dead Letter Queue (tabJobFailed)
			logger.Log.Error("Job PERMANENTLY FAILED. Moving to DLQ.", zap.Uint("job_id", j.ID))
			database.DB.Table("tabJobFailed").Create(j)
			database.DB.Table("tabJob").Where("id = ?", j.ID).Delete(&Job{})
		} else {
			database.DB.Table("tabJob").Where("id = ?", j.ID).Updates(map[string]interface{}{
				"status":      status,
				"error_log":   err.Error(),
				"retry_count": newRetryCount,
				"locked_at":   nil, // Release for other nodes
				"updated_at":  nextRetryAt, // Schedule next attempt
			})
		}
	} else {
		updateJobStatus(j.ID, Completed, "")
	}
}

	}
}

func updateJobStatus(id uint, status JobStatus, errorLog string) {
	database.DB.Table("tabJob").Where("id = ?", id).Updates(map[string]interface{}{
		"status":    status,
		"error_log": errorLog,
	})
}
