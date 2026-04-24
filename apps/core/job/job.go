package job

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goerp/goerp/apps/core/logger"
	"github.com/hibiken/asynq"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var Client *asynq.Client
var Server *asynq.Server
var Mux *asynq.ServeMux

func InitJob() {
	redisAddr := fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port"))
	redisConn := asynq.RedisClientOpt{Addr: redisAddr}

	Client = asynq.NewClient(redisConn)
	Server = asynq.NewServer(redisConn, asynq.Config{
		Concurrency: 10,
	})
	Mux = asynq.NewServeMux()

	logger.Log.Info("Job queue initialized", zap.String("redis", redisAddr))
}

func Enqueue(typeName string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(typeName, data)
	info, err := Client.Enqueue(task)
	if err != nil {
		return err
	}

	logger.Log.Debug("Enqueued job", zap.String("id", info.ID), zap.String("type", typeName))
	return nil
}

func StartWorker() {
	if err := Server.Run(Mux); err != nil {
		logger.Log.Fatal("Job worker failed", zap.Error(err))
	}
}

// Helper to register handler
func RegisterHandler(typeName string, handler func(context.Context, *asynq.Task) error) {
	Mux.HandleFunc(typeName, handler)
}
