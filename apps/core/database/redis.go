package database

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goerp/goerp/apps/core/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var Redis *redis.Client
var Ctx = context.Background()

func InitRedis() {
	addr := fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port"))
	Redis = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})

	if err := Redis.Ping(Ctx).Err(); err != nil {
		logger.Log.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	logger.Log.Info("Redis connection established", zap.String("addr", addr))
}
