package datastore

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type RedisDatastore struct {
	client *redis.Client
}

const (
	addr = "localhost:6379"
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

// LRem 最初の count 個だけ value にマッチする要素を key に対応するリストから削除する。 count が負数の場合は最後から count 個だけ削除する。
func (rds *RedisDatastore) LRem(ctx context.Context, key string, count int64, value string) error {
	return rds.LRem(ctx, key, count, value)
}

// LRange: start <= x <= end を満たす要素を取得する。 (0-indexed)
func (rds *RedisDatastore) LRange(ctx context.Context, key string, start int64, end int64) ([]string, error) {
	return rds.client.LRange(ctx, key, start, end).Result()
}

func (rds *RedisDatastore) LastPop(ctx context.Context, key string) (string, error) {
	return rds.client.RPop(ctx, key).Result()
}

func (rds *RedisDatastore) LastPush(ctx context.Context, key string, value string) (int64, error) {
	return rds.client.RPush(ctx, key, value).Result()
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

func (rds *RedisDatastore) Insert(ctx context.Context, key string, score float64, member interface{}) error {
	return rds.client.ZAdd(ctx, key, &redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

// GetMinElement: SortedSetの最小値を返す。
func (rds *RedisDatastore) GetMinElement(ctx context.Context, key string) (string, error) {
	// 最小値を取り出して再度代入する。
	val, err := rds.client.ZPopMin(ctx, key).Result()
	if err != nil {
		return "", err
	}
	if len(val) == 0 {
		return "", errors.New("failed to get min element. its empty. (GetMinElement)")
	}
	x, ok := val[0].Member.(string)
	if !ok {
		return "", errors.New("failed to convert string. (GetMinElement)")
	}
	if err := rds.Insert(ctx, key, val[0].Score, val[0].Member); err != nil {
		return x, err
	}
	return x, nil
}

// TODO: hMSet メソッドを利用することでより効率的に書けるらしい
// SetHash: 親Key 子Key Valueにより管理されるデータ型に値を格納する。
func (rds *RedisDatastore) SetHash(ctx context.Context, pkey string, ckey string, value string) error {
	return rds.client.HSet(ctx, pkey, ckey, value).Err()
}

// GetHash: 親Key, 子Key で値を取得する
func (rds *RedisDatastore) GetHash(ctx context.Context, pkey string, ckey string) (string, error) {
	return rds.client.HGet(ctx, pkey, ckey).Result()
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
