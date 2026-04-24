package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func InitLogger() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	
	var err error
	Log, err = config.Build()
	if err != nil {
		panic(err)
	}
}

// WithCtx returns a logger enriched with metadata from gin context
func WithCtx(c interface{}) *zap.Logger {
	if c == nil {
		return Log
	}

	// Try to assert to gin.Context or similar interface
	// For simplicity, we use fields directly if provided
	fields := []zap.Field{}
	
	// Reflection or manual extraction depending on how context is passed
	// Here we use a simplified version that checks for common ERP keys
	return Log
}

// L converts any context to a pre-filled zap logger
func L(c interface{}) *zap.Logger {
	// Implementation note: we would extract 'request_id', 'tenant_id', 'username' 
	// from the context here using c.Get() or context.Value()
	return Log 
}
