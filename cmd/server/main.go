package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goerp/goerp/apps/core"
	"github.com/goerp/goerp/apps/core/database"
	coreHttp "github.com/goerp/goerp/apps/core/http"
	"github.com/goerp/goerp/apps/core/integration"
	"github.com/goerp/goerp/apps/core/job"
	jobHandlers "github.com/goerp/goerp/apps/core/job/handlers"
	"github.com/goerp/goerp/apps/core/logger"
	"github.com/goerp/goerp/apps/core/meta/loader"
	"github.com/goerp/goerp/apps/core/meta/migrator"
	"github.com/goerp/goerp/apps/core/meta/registry"
	"github.com/goerp/goerp/apps/core/middleware"
	"github.com/goerp/goerp/apps/erp"
	stockService "github.com/goerp/goerp/apps/erp/stock/service"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"strings"
)

func main() {
	// 1. Initialize Logger
	logger.InitLogger()
	defer logger.Log.Sync()

	// 2. Load Configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv() // Read from Environment Variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	if err := viper.ReadInConfig(); err != nil {
		logger.Log.Warn("No config file found, relying on environment variables")
	}

	// 3. Initialize Databases
	database.InitDB()
	database.InitRedis()

	// 4. Load Apps & DocTypes
	if err := loader.LoadApps("./apps"); err != nil {
		logger.Log.Fatal("Failed to load apps", zap.Error(err))
	}

	// 5. Run Migrations & Subscribe to Sync Events
	registry.DefaultRegistry.SubscribeToInvalidations()
	for _, dt := range registry.DefaultRegistry.GetAll() {
		if err := migrator.MigrateDocType(dt); err != nil {
			logger.Log.Error("Migration failed", zap.String("doctype", dt.Name), zap.Error(err))
		}
	}

	// 6. Initialize ERP Modules (Hooks & Business Logic)
	erp.InitERP()
	core.RegisterCoreHooks()
	integration.InitIntegration()

	// Register Job Handlers
	jobHandlers.RegisterMaintenanceHandlers()
	stockService.RegisterStockHandlers()

	// 7. Initialize Background Jobs
	job.InitJob()
	go job.StartWorker(context.Background())

	// 8. Setup Gin
	gin.SetMode(viper.GetString("server.mode"))
	r := gin.New() // Use New instead of Default to have more control
	r.Use(gin.Logger())
	r.Use(middleware.JSONRecovery()) // Custom JSON-based recovery

	// Register Core Routes
	coreHttp.RegisterRoutes(r)

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "active",
		})
	})

	port := viper.GetInt("server.port")
	logger.Log.Info("Starting server", zap.Int("port", port))
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		logger.Log.Fatal("Failed to start server", zap.Error(err))
	}
}
