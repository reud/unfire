package persistence

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type RedisDatastore struct {
	client *redis.Client
}

const (
	addr = "127.0.0.1:6379"
)

func (rds *RedisDatastore) SetString(ctx context.Context, key string, value string) error {
	return rds.client.Set(ctx, key, value, 0).Err()
}

func (rds *RedisDatastore) GetString(ctx context.Context, key string) (string, error) {
	return rds.client.Get(ctx, key).Result()
}

func (rds *RedisDatastore) AppendString(ctx context.Context, key string, value string) error {
	return rds.client.RPush(ctx, key, value).Err()
}

func (rds *RedisDatastore) LastPop(ctx context.Context, key string) (string, error) {
	return rds.client.RPop(ctx, key).Result()
}

func (rds *RedisDatastore) ListLen(ctx context.Context, key string) (int64, error) {
	return rds.client.LLen(ctx, key).Result()
}

func (rds *RedisDatastore) GetStringByIndex(ctx context.Context, key string, index int64) (string, error) {
	return rds.client.LIndex(ctx, key, index).Result()
}

func (rds *RedisDatastore) InsertInt64(ctx context.Context, key string, value int64) error {
	return rds.client.ZAdd(ctx, key, &redis.Z{
		Score:  float64(value - 9007199254740992),
		Member: value,
	}).Err()
}

// GetMinElement: SortedSetの最小値を返す。
func (rds *RedisDatastore) GetMinElement(ctx context.Context, key string) (int64, error) {
	// 最小値を取り出して再度代入する。
	val, err := rds.client.ZPopMin(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	x, ok := val[0].Member.(int64)
	if !ok {
		return 0, errors.New("failed to convert int64. (GetMinElement)")
	}
	if err := rds.InsertInt64(ctx, key, x); err != nil {
		return x, errors.Wrap(err, "failed to reinsert element (GetMinElement)")
	}
	return x, nil
}

// PopMin: SortedSetの最小値を削除する。
func (rds *RedisDatastore) PopMin(ctx context.Context, key string) error {
	return rds.client.ZPopMin(ctx, key).Err()
}

func NewRedisDatastore() (Datastore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	return &RedisDatastore{client: client}, nil
}