package loader

import (
	"encoding/json"
	"fmt"
	"github.com/goerp/goerp/apps/core/logger"
	"github.com/goerp/goerp/apps/core/meta/registry"
	"github.com/goerp/goerp/apps/core/meta/types"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
)

func LoadApps(appsDir string) error {
	dirs := []string{appsDir, "./plugins"} // Scan both standard apps and external plugins

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue // Skip if directory doesn't exist
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				appPath := filepath.Join(dir, entry.Name())
				if err := loadApp(appPath); err != nil {
					logger.Log.Error("Failed to load app/plugin", zap.String("path", appPath), zap.Error(err))
				}
			}
		}
	}
	return nil
}

func loadApp(appPath string) error {
	modulesFile := filepath.Join(appPath, "modules.txt")
	content, err := os.ReadFile(modulesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // App might not have modules.txt
		}
		return err
	}

	modules := strings.Split(string(content), "\n")
	for _, moduleName := range modules {
		moduleName = strings.TrimSpace(moduleName)
		if moduleName == "" {
			continue
		}

		modulePath := filepath.Join(appPath, moduleName)
		if err := loadModule(modulePath, moduleName); err != nil {
			logger.Log.Error("Failed to load module", zap.String("module", moduleName), zap.Error(err))
		}
	}

	return nil
}

func loadModule(modulePath string, moduleName string) error {
	doctypeDir := filepath.Join(modulePath, "doctype")
	entries, err := os.ReadDir(doctypeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			schemaFile := filepath.Join(doctypeDir, entry.Name(), "schema.json")
			if err := loadDocType(schemaFile); err != nil {
				logger.Log.Error("Failed to load doctype", zap.String("path", schemaFile), zap.Error(err))
			}
		}
	}

	return nil
}

func loadDocType(schemaFile string) error {
	content, err := os.ReadFile(schemaFile)
	if err != nil {
		return err
	}

	var dt types.DocType
	if err := json.Unmarshal(content, &dt); err != nil {
		return fmt.Errorf("invalid schema json: %w", err)
	}

	if err := registry.DefaultRegistry.Register(&dt); err != nil {
		return err
	}

	logger.Log.Info("Registered DocType", zap.String("name", dt.Name), zap.String("module", dt.Module))
	return nil
}
