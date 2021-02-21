package datastore

import "context"

// KeyValueストア
type Datastore interface {
	SetString(ctx context.Context, key string, value string) error
	GetString(ctx context.Context, key string) (string, error)
	// List出ない場合はエラー
	AppendString(ctx context.Context, key string, value string) error
	// 末尾の要素を削除して削除した要素を返す(Listでないならエラー)
	LastPop(ctx context.Context, key string) (string, error)
	// 末尾に要素を追加する。(Listでないならエラー)
	LastPush(ctx context.Context, key string, value string) (int64, error)
	// KeyのListの長さを返す(Listでないならエラー)
	ListLen(ctx context.Context, key string) (int64, error)
	// GetStringByIndex Stringをindex指定で返す。
	GetStringByIndex(ctx context.Context, key string, index int64) (string, error)
	// SortedSetにint64を追加する (score == value)
	InsertInt64(ctx context.Context, key string, value int64) error
	// SortedSetの最小値を取得する。
	GetMinElement(ctx context.Context, key string) (string, error)
	// PopMin: SortedSetの最小値を削除する。
	PopMin(ctx context.Context, key string) error
	// SortedSetに任意の値を追加する。
	Insert(ctx context.Context, key string, score float64, member interface{}) error
	SetHash(ctx context.Context, pkey string, ckey string, value string) error
	GetHash(ctx context.Context, pkey string, ckey string) (string, error)
	// LRange: start <= x <= end を満たす要素を取得する。 (0-indexed)
	LRange(ctx context.Context, key string, start int64, end int64) ([]string, error)
	// LRem: 最初の count 個だけ value にマッチする要素を key に対応するリストから削除する。 count が負数の場合は最後から count 個だけ削除する。
	LRem(ctx context.Context, key string, count int64, value string) error
}
