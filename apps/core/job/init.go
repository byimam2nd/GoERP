package job

import (
	"context"
	"github.com/goerp/goerp/apps/core/database"
	"time"
)

// InitJob menginisialisasi tabel job dan mendaftarkan scheduler dasar.
func InitJob() {
	// 1. Auto Migrate tabel tabJob
	database.DB.AutoMigrate(&Job{})

	// 2. Daftarkan cron job sederhana untuk maintenance (misal: setiap jam)
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			Enqueue("system", "PartitionMaintenance", map[string]interface{}{
				"scheduled_at": time.Now(),
			})
		}
	}()
}
