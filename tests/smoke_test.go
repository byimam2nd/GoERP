package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goerp/goerp/apps/core/database"
	coreHttp "github.com/goerp/goerp/apps/core/http"
	"github.com/goerp/goerp/apps/core/meta/loader"
	"github.com/goerp/goerp/apps/core/meta/migrator"
	"github.com/goerp/goerp/apps/core/meta/registry"
	"github.com/goerp/goerp/apps/core/multitenant"
	"github.com/goerp/goerp/apps/core/logger"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestDB() {
	logger.InitLogger()
	viper.Set("server.secret_key", "test_secret")
	
	// Use SQLite In-Memory for testing
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	database.DB = db
}

func TestFullFlow(t *testing.T) {
	setupTestDB()
	
	// 1. Load Apps (Assume we are running from project root)
	err := loader.LoadApps("../apps")
	require.NoError(t, err)

	// 2. Run Migrations
	for _, dt := range registry.DefaultRegistry.GetAll() {
		err := migrator.MigrateDocType(dt)
		require.NoError(t, err)
	}

	// 3. Provision Tenant
	err = multitenant.ProvisionTenant("Test Corp", "test.goerp.com", "admin@test.com", "password123")
	require.NoError(t, err)

	// Get the generated tenant ID
	var tenant struct {
		Name string
	}
	err = database.DB.Table("tabTenant").First(&tenant).Error
	require.NoError(t, err)
	tenantID := tenant.Name

	// 4. Setup Router
	gin.SetMode(gin.TestMode)
	r := gin.New()
	coreHttp.RegisterRoutes(r)

	// 5. Test Login
	loginKey := map[string]string{
		"username": "admin",
		"password": "password123",
	}
	loginJSON, _ := json.Marshal(loginKey)
	
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GoERP-Tenant", tenantID)
	
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, 200, w.Code)
	
	var loginResp map[string]string
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	token := loginResp["token"]
	assert.NotEmpty(t, token)

	// 6. Test Create Company (CRUD)
	companyData := map[string]interface{}{
		"company_name":     "Test Company LTD",
		"abbr":             "TCL",
		"default_currency": "IDR",
	}
	companyJSON, _ := json.Marshal(companyData)
	
	req, _ = http.NewRequest("POST", "/api/v1/resource/Company", bytes.NewBuffer(companyJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GoERP-Tenant", tenantID)
	req.Header.Set("Content-Type", "application/json")
	
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, 200, w.Code)
	fmt.Println("Integration Test Passed: Tenant Provisioned, Authenticated, and Data Created.")
}
