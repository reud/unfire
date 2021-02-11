package scenario

import (
	"context"
	"testing"
	"time"
	"unfire/domain/service"
	"unfire/infrastructure/datastore"
	"unfire/usecase/batch"
	"unfire/usecase/handler"
	"unfire/usecase/handler/admin"

	"github.com/go-redis/redis/v8"
)

type UseCases struct {
	Au handler.AuthUseCase
	Ru admin.RestartUseCase
}

type Scenario struct {
	Work func(t *testing.T, cases UseCases)
}

// テスト準備
func flushDB() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	client.FlushDB(context.Background())
}

// 各ケース毎に利用する。
func initialize() batch.Services {
	flushDB()
	ds, err := datastore.NewRedisDatastore()
	if err != nil {
		panic(err)
	}
	dc := service.NewDatastoreController(ds)

	// start reload batch
	rbth := batch.NewReloadBatchService(time.Minute*3, dc)
	dbth := batch.NewDeleteBatchService(time.Minute*3, dc)
	batchSv := batch.NewServices(rbth, dbth)
	return batchSv
}
